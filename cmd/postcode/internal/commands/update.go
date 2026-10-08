package commands

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultRepoOwner       = "abcubed3"
	defaultRepoName        = "postcode"
	defaultGitHubAPIURL    = "https://api.github.com"
	fallbackInstallCmd     = "go install github.com/abcubed3/postcode/cmd/postcode@latest"
	maxArchiveDownloadSize = 250 * 1024 * 1024 // 250 MB max download size to prevent resource exhaustion
	maxBinaryPayloadSize   = 150 * 1024 * 1024 // 150 MB max unpacked binary size
)

// GitHubRelease represents release metadata from GitHub API.
type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	PublishedAt string        `json:"published_at"`
	HTMLURL     string        `json:"html_url"`
	Assets      []GitHubAsset `json:"assets"`
}

// GitHubAsset represents a downloadable release asset file.
type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	ContentType        string `json:"content_type"`
}

// UpdateCheckResult holds the status of a release version comparison.
type UpdateCheckResult struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
	ReleaseTag      string `json:"release_tag"`
	ReleaseURL      string `json:"release_url"`
	PublishedAt     string `json:"published_at,omitempty"`
	StatusMessage   string `json:"status_message"`
}

func (r UpdateCheckResult) CSVHeader() []string {
	return []string{"current_version", "latest_version", "update_available", "release_tag", "release_url"}
}

func (r UpdateCheckResult) CSVRows() [][]string {
	return [][]string{{
		r.CurrentVersion,
		r.LatestVersion,
		fmt.Sprintf("%t", r.UpdateAvailable),
		r.ReleaseTag,
		r.ReleaseURL,
	}}
}

// UpdateResult captures the outcome of an in-place upgrade.
type UpdateResult struct {
	PreviousVersion string `json:"previous_version"`
	UpdatedVersion  string `json:"updated_version"`
	BinaryPath      string `json:"binary_path"`
	Success         bool   `json:"success"`
	Message         string `json:"message"`
}

func (r UpdateResult) CSVHeader() []string {
	return []string{"previous_version", "updated_version", "binary_path", "success", "message"}
}

func (r UpdateResult) CSVRows() [][]string {
	return [][]string{{
		r.PreviousVersion,
		r.UpdatedVersion,
		r.BinaryPath,
		fmt.Sprintf("%t", r.Success),
		r.Message,
	}}
}

// UpdaterConfig defines parameters for discovering and performing self-updates.
type UpdaterConfig struct {
	CurrentVersion   string
	RepoOwner        string
	RepoName         string
	GitHubAPIURL     string
	GitHubToken      string
	TargetExecutable string
	GOOS             string
	GOARCH           string
	HTTPClient       *http.Client
}

// Updater manages checking and executing self-updates.
type Updater struct {
	cfg UpdaterConfig
}

var defaultHTTPTransport http.RoundTripper

// SetDefaultHTTPTransport sets a package-level transport for in-memory testing without network calls.
func SetDefaultHTTPTransport(rt http.RoundTripper) {
	defaultHTTPTransport = rt
}

// NewUpdater initializes a new Updater instance.
func NewUpdater(cfg UpdaterConfig) *Updater {
	if cfg.CurrentVersion == "" {
		cfg.CurrentVersion = Version
	}
	if cfg.RepoOwner == "" {
		cfg.RepoOwner = defaultRepoOwner
	}
	if cfg.RepoName == "" {
		cfg.RepoName = defaultRepoName
	}
	if cfg.GitHubAPIURL == "" {
		cfg.GitHubAPIURL = defaultGitHubAPIURL
	}
	if cfg.GOOS == "" {
		cfg.GOOS = runtime.GOOS
	}
	if cfg.GOARCH == "" {
		cfg.GOARCH = runtime.GOARCH
	}
	if cfg.HTTPClient == nil {
		client := &http.Client{Timeout: 30 * time.Second}
		if defaultHTTPTransport != nil {
			client.Transport = defaultHTTPTransport
		}
		cfg.HTTPClient = client
	} else if cfg.HTTPClient.Transport == nil && defaultHTTPTransport != nil {
		cfg.HTTPClient.Transport = defaultHTTPTransport
	}
	return &Updater{cfg: cfg}
}

// FetchLatestRelease queries GitHub Releases API for the latest release metadata.
func (u *Updater) FetchLatestRelease(ctx context.Context) (*GitHubRelease, error) {
	apiURL := strings.TrimRight(u.cfg.GitHubAPIURL, "/")
	var endpoint string
	if strings.Contains(apiURL, "/releases/") {
		endpoint = apiURL
	} else {
		endpoint = fmt.Sprintf("%s/repos/%s/%s/releases/latest", apiURL, u.cfg.RepoOwner, u.cfg.RepoName)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "postcode-cli-updater")
	if token := u.getAuthToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := u.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching latest release from GitHub: %w\nManual update fallback:\n    %s", err, fallbackInstallCmd)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound && !strings.Contains(apiURL, "/releases/") {
		// Fallback: If /releases/latest returns 404 (e.g. repo only has prereleases), try /releases list
		fallbackEndpoint := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=10", apiURL, u.cfg.RepoOwner, u.cfg.RepoName)
		fbReq, fbErr := http.NewRequestWithContext(ctx, http.MethodGet, fallbackEndpoint, nil)
		if fbErr == nil {
			fbReq.Header.Set("Accept", "application/vnd.github.v3+json")
			fbReq.Header.Set("User-Agent", "postcode-cli-updater")
			if token := u.getAuthToken(); token != "" {
				fbReq.Header.Set("Authorization", "Bearer "+token)
			}
			fbResp, doErr := u.cfg.HTTPClient.Do(fbReq)
			if doErr == nil {
				defer fbResp.Body.Close()
				if fbResp.StatusCode == http.StatusOK {
					var releases []GitHubRelease
					if err := json.NewDecoder(fbResp.Body).Decode(&releases); err == nil {
						for _, r := range releases {
							if !r.Draft {
								return &r, nil
							}
						}
					}
				} else {
					fbBody, _ := io.ReadAll(io.LimitReader(fbResp.Body, 1024))
					isFbRateLimit := (fbResp.StatusCode == http.StatusForbidden || fbResp.StatusCode == http.StatusTooManyRequests) &&
						(fbResp.Header.Get("X-RateLimit-Remaining") == "0" || strings.Contains(strings.ToLower(string(fbBody)), "rate limit"))
					if isFbRateLimit {
						return nil, fmt.Errorf("GitHub API rate limit exceeded. Set GITHUB_TOKEN environment variable or install manually via:\n    %s", fallbackInstallCmd)
					}
				}
			}
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		isRateLimit := (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests) &&
			(resp.Header.Get("X-RateLimit-Remaining") == "0" || strings.Contains(strings.ToLower(string(body)), "rate limit"))
		if isRateLimit {
			return nil, fmt.Errorf("GitHub API rate limit exceeded. Set GITHUB_TOKEN environment variable or install manually via:\n    %s", fallbackInstallCmd)
		}
		return nil, fmt.Errorf("GitHub API returned HTTP %d: %s\nManual update fallback:\n    %s", resp.StatusCode, strings.TrimSpace(string(body)), fallbackInstallCmd)
	}

	var rel GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("parsing GitHub release response: %w\nManual update fallback:\n    %s", err, fallbackInstallCmd)
	}

	return &rel, nil
}

func (u *Updater) getAuthToken() string {
	var token string
	if u.cfg.GitHubToken != "" {
		token = u.cfg.GitHubToken
	} else if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
		token = envToken
	} else if envToken := os.Getenv("POSTCODE_GITHUB_TOKEN"); envToken != "" {
		token = envToken
	}
	token = strings.TrimSpace(token)
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimPrefix(token, "token ")
	return strings.TrimSpace(token)
}

// Check queries for updates and compares against the running version without modifying local files.
func (u *Updater) Check(ctx context.Context) (*UpdateCheckResult, *GitHubRelease, error) {
	rel, err := u.FetchLatestRelease(ctx)
	if err != nil {
		return nil, nil, err
	}

	current := u.cfg.CurrentVersion
	latest := rel.TagName

	cmp := CompareVersions(current, latest)
	updateAvailable := cmp < 0

	var msg string
	if updateAvailable {
		msg = fmt.Sprintf("A newer version (%s) is available. You are running %s.", latest, current)
	} else if cmp == 0 {
		msg = fmt.Sprintf("postcode is already up-to-date (%s).", current)
	} else {
		msg = fmt.Sprintf("postcode is up-to-date (running %s, latest release is %s).", current, latest)
	}

	res := &UpdateCheckResult{
		CurrentVersion:  current,
		LatestVersion:   latest,
		UpdateAvailable: updateAvailable,
		ReleaseTag:      rel.TagName,
		ReleaseURL:      rel.HTMLURL,
		PublishedAt:     rel.PublishedAt,
		StatusMessage:   msg,
	}

	return res, rel, nil
}

// Upgrade downloads matching assets, verifies checksums, and replaces the local binary.
func (u *Updater) Upgrade(ctx context.Context, rel *GitHubRelease, progressWriter io.Writer) (*UpdateResult, error) {
	if rel == nil {
		var err error
		rel, err = u.FetchLatestRelease(ctx)
		if err != nil {
			return nil, err
		}
	}

	// 1. Locate target binary
	targetPath := u.cfg.TargetExecutable
	if targetPath == "" {
		execPath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("determining executable path: %w\nManual update fallback:\n    %s", err, fallbackInstallCmd)
		}
		realPath, err := filepath.EvalSymlinks(execPath)
		if err == nil && realPath != "" {
			targetPath = realPath
		} else {
			targetPath = execPath
		}
	}

	// 2. Identify expected release archive
	archiveName, binNameInArchive := getExpectedAssetNames(u.cfg.GOOS, u.cfg.GOARCH)

	// 3. Locate matching asset and checksums.txt
	var archiveAsset *GitHubAsset
	var checksumAsset *GitHubAsset

	for i := range rel.Assets {
		asset := &rel.Assets[i]
		if strings.EqualFold(asset.Name, archiveName) {
			archiveAsset = asset
		} else if strings.EqualFold(asset.Name, "checksums.txt") ||
			strings.EqualFold(asset.Name, "sha256sums") ||
			strings.EqualFold(asset.Name, "checksums.sha256") ||
			strings.EqualFold(asset.Name, "sha256sums.txt") ||
			strings.EqualFold(asset.Name, "checksums.sha256.txt") {
			if checksumAsset == nil || strings.EqualFold(asset.Name, "checksums.txt") {
				checksumAsset = asset
			}
		}
	}

	// Fallback asset matching: locate archive matching target OS and arch
	if archiveAsset == nil {
		for i := range rel.Assets {
			asset := &rel.Assets[i]
			nameLower := strings.ToLower(asset.Name)
			if strings.HasPrefix(nameLower, "postcode") &&
				strings.Contains(nameLower, strings.ToLower(u.cfg.GOOS)) &&
				strings.Contains(nameLower, strings.ToLower(u.cfg.GOARCH)) {
				if (u.cfg.GOOS == "windows" && strings.HasSuffix(nameLower, ".zip")) ||
					(u.cfg.GOOS != "windows" && (strings.HasSuffix(nameLower, ".tar.gz") || strings.HasSuffix(nameLower, ".tgz"))) {
					archiveAsset = asset
					break
				}
			}
		}
	}

	if archiveAsset == nil {
		return nil, fmt.Errorf("no pre-compiled binary available for %s/%s (expected %q) in release %s.\nAutomatic binary upgrade is not supported on this platform.\nTo update, run:\n    %s",
			u.cfg.GOOS, u.cfg.GOARCH, archiveName, rel.TagName, fallbackInstallCmd)
	}

	if checksumAsset == nil {
		return nil, fmt.Errorf("checksums.txt asset not found in release %s.\nFor security, updates cannot proceed without published checksums.\nTo update, run:\n    %s",
			rel.TagName, fallbackInstallCmd)
	}

	// 4. Download checksums.txt
	if progressWriter != nil {
		_, _ = fmt.Fprintf(progressWriter, "Fetching checksums from %s...\n", checksumAsset.Name)
	}
	checksumsData, err := u.downloadBytes(ctx, checksumAsset.BrowserDownloadURL)
	if err != nil {
		return nil, fmt.Errorf("downloading checksums.txt: %w\nManual update fallback:\n    %s", err, fallbackInstallCmd)
	}
	checksumMap := parseChecksums(string(checksumsData))

	expectedHash, ok := checksumMap[strings.ToLower(archiveAsset.Name)]
	if !ok {
		expectedHash, ok = checksumMap[strings.ToLower(archiveName)]
	}
	if !ok {
		return nil, fmt.Errorf("checksum for %s not found in release checksums.txt.\nTo update, run:\n    %s", archiveAsset.Name, fallbackInstallCmd)
	}

	// 5. Download binary archive
	if progressWriter != nil {
		_, _ = fmt.Fprintf(progressWriter, "Downloading release archive %s...\n", archiveAsset.Name)
	}
	archiveBytes, err := u.downloadBytes(ctx, archiveAsset.BrowserDownloadURL)
	if err != nil {
		return nil, fmt.Errorf("downloading release archive %s: %w\nManual update fallback:\n    %s", archiveAsset.Name, err, fallbackInstallCmd)
	}

	// 6. Verify checksum
	if progressWriter != nil {
		_, _ = fmt.Fprintln(progressWriter, "Verifying archive checksum...")
	}
	hasher := sha256.New()
	hasher.Write(archiveBytes)
	computedHash := hex.EncodeToString(hasher.Sum(nil))

	if !strings.EqualFold(computedHash, expectedHash) {
		return nil, fmt.Errorf("checksum verification failed for %s: expected %s, got %s.\nTo update, run:\n    %s", archiveAsset.Name, expectedHash, computedHash, fallbackInstallCmd)
	}
	if progressWriter != nil {
		_, _ = fmt.Fprintf(progressWriter, "Checksum verified: %s\n", computedHash)
	}

	// 7. Extract binary from archive
	if progressWriter != nil {
		_, _ = fmt.Fprintln(progressWriter, "Extracting binary payload...")
	}
	var newBinary []byte
	if strings.HasSuffix(strings.ToLower(archiveName), ".tar.gz") {
		newBinary, err = extractBinaryFromTarGz(bytes.NewReader(archiveBytes), binNameInArchive)
	} else if strings.HasSuffix(strings.ToLower(archiveName), ".zip") {
		newBinary, err = extractBinaryFromZip(archiveBytes, binNameInArchive)
	} else {
		return nil, fmt.Errorf("unsupported archive format: %s\nManual update fallback:\n    %s", archiveName, fallbackInstallCmd)
	}
	if err != nil {
		return nil, fmt.Errorf("extracting binary from archive: %w\nManual update fallback:\n    %s", err, fallbackInstallCmd)
	}
	if len(newBinary) == 0 {
		return nil, fmt.Errorf("extracted binary is empty\nManual update fallback:\n    %s", fallbackInstallCmd)
	}

	// 8. Atomically replace target binary in-place with rollback
	if progressWriter != nil {
		_, _ = fmt.Fprintf(progressWriter, "Installing new binary to %s...\n", targetPath)
	}
	if err := replaceBinary(targetPath, newBinary); err != nil {
		return nil, err
	}

	return &UpdateResult{
		PreviousVersion: u.cfg.CurrentVersion,
		UpdatedVersion:  rel.TagName,
		BinaryPath:      targetPath,
		Success:         true,
		Message:         fmt.Sprintf("Successfully upgraded postcode to %s", rel.TagName),
	}, nil
}

func (u *Updater) downloadBytes(ctx context.Context, downloadURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "postcode-cli-updater")
	resp, err := u.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		if len(body) > 0 {
			return nil, fmt.Errorf("HTTP %d: %s (%s)", resp.StatusCode, resp.Status, strings.TrimSpace(string(body)))
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxArchiveDownloadSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxArchiveDownloadSize {
		return nil, fmt.Errorf("downloaded archive exceeds maximum size limit of %d bytes", maxArchiveDownloadSize)
	}
	return data, nil
}

// getExpectedAssetNames returns standard release asset archive and binary names for the target OS/Arch.
func getExpectedAssetNames(goos, goarch string) (archiveName string, binName string) {
	if goos == "windows" {
		archiveName = fmt.Sprintf("postcode-windows-%s.zip", goarch)
		binName = fmt.Sprintf("postcode-windows-%s.exe", goarch)
	} else {
		archiveName = fmt.Sprintf("postcode-%s-%s.tar.gz", goos, goarch)
		binName = fmt.Sprintf("postcode-%s-%s", goos, goarch)
	}
	return archiveName, binName
}

// parseChecksums extracts a filename-to-sha256 map from checksums.txt content.
// Supports both GNU format (<hash>  <filename>) and BSD format (SHA256 (<filename>) = <hash>).
func parseChecksums(content string) map[string]string {
	checksums := make(map[string]string)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Support BSD format: SHA256 (filename) = hash
		lowerLine := strings.ToLower(line)
		if strings.HasPrefix(lowerLine, "sha256") || strings.HasPrefix(lowerLine, "sha-256") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				namePart := strings.TrimSpace(parts[0])
				start := strings.Index(namePart, "(")
				end := strings.LastIndex(namePart, ")")
				if start != -1 && end != -1 && end > start {
					rawName := namePart[start+1 : end]
					name := path.Base(strings.ReplaceAll(strings.TrimSpace(rawName), "\\", "/"))
					hash := strings.ToLower(strings.TrimSpace(parts[1]))
					if name != "" && len(hash) == 64 {
						checksums[strings.ToLower(name)] = hash
						continue
					}
				}
			}
		}

		// Support GNU format: <hash>  <filename> or <hash> *<filename>
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			hash := strings.ToLower(fields[0])
			if len(hash) == 64 {
				rest := strings.TrimSpace(line[len(fields[0]):])
				rest = strings.TrimPrefix(rest, "*")
				name := path.Base(strings.ReplaceAll(strings.TrimSpace(rest), "\\", "/"))
				if name != "" {
					checksums[strings.ToLower(name)] = hash
					continue
				}
			}
			// Fallback: <filename>  <hash>
			lastHash := strings.ToLower(fields[len(fields)-1])
			if len(lastHash) == 64 {
				rest := strings.TrimSpace(line[:len(line)-len(fields[len(fields)-1])])
				rest = strings.TrimPrefix(rest, "*")
				name := path.Base(strings.ReplaceAll(strings.TrimSpace(rest), "\\", "/"))
				if name != "" {
					checksums[strings.ToLower(name)] = lastHash
					continue
				}
			}
		}
	}
	return checksums
}

var nonExecutableExtensions = map[string]bool{
	".txt":      true,
	".md":       true,
	".markdown": true,
	".rst":      true,
	".rtf":      true,
	".json":     true,
	".sha256":   true,
	".sig":      true,
	".yaml":     true,
	".yml":      true,
	".toml":     true,
	".xml":      true,
	".html":     true,
	".htm":      true,
	".css":      true,
	".js":       true,
	".ts":       true,
	".tsx":      true,
	".jsx":      true,
	".map":      true,
	".sh":       true,
	".bash":     true,
	".zsh":      true,
	".fish":     true,
	".bat":      true,
	".cmd":      true,
	".ps1":      true,
	".py":       true,
	".rb":       true,
	".pl":       true,
	".png":      true,
	".jpg":      true,
	".jpeg":     true,
	".gif":      true,
	".svg":      true,
	".ico":      true,
	".pdf":      true,
	".doc":      true,
	".docx":     true,
	".license":  true,
	".lic":      true,
	".bak":      true,
	".old":      true,
	".tmp":      true,
	".1":        true,
	".man":      true,
	".tar":      true,
	".gz":       true,
	".tgz":      true,
	".zip":      true,
	".manifest": true,
	".service":  true,
	".env":      true,
	".wasm":     true,
	".mod":      true,
	".sum":      true,
	".log":      true,
	".csv":      true,
	".tsv":      true,
	".sql":      true,
	".ini":      true,
	".conf":     true,
	".cfg":      true,
	".pub":      true,
	".key":      true,
	".crt":      true,
	".pem":      true,
	".cer":      true,
	".rpm":      true,
	".deb":      true,
	".apk":      true,
	".dmg":      true,
	".pkg":      true,
}

// isExecutableCandidate evaluates whether an archive entry represents an executable file.
func isExecutableCandidate(base, targetName string) (isExact bool, isCandidate bool) {
	cleanBase := strings.ReplaceAll(base, "\\", "/")
	cleanBase = path.Base(cleanBase)
	lowerBase := strings.ToLower(cleanBase)
	ext := strings.ToLower(filepath.Ext(cleanBase))
	if nonExecutableExtensions[ext] {
		return false, false
	}

	// Reject auxiliary documents and test artifacts without extensions
	if strings.Contains(lowerBase, "license") ||
		strings.Contains(lowerBase, "readme") ||
		strings.Contains(lowerBase, "completion") ||
		strings.Contains(lowerBase, "changelog") ||
		strings.Contains(lowerBase, "copying") ||
		strings.Contains(lowerBase, "notice") ||
		strings.Contains(lowerBase, "authors") ||
		strings.Contains(lowerBase, "contributing") ||
		strings.Contains(lowerBase, "test") ||
		strings.Contains(lowerBase, "bench") ||
		strings.Contains(lowerBase, "example") ||
		strings.Contains(lowerBase, "sample") {
		return false, false
	}

	// If looking for a Windows executable, require .exe extension
	if strings.HasSuffix(strings.ToLower(targetName), ".exe") && !strings.HasSuffix(lowerBase, ".exe") {
		return false, false
	}

	// Exact matches have highest priority
	targetTrimmed := strings.TrimSuffix(targetName, ".exe")
	if strings.EqualFold(cleanBase, targetName) ||
		strings.EqualFold(cleanBase, targetTrimmed) ||
		strings.EqualFold(cleanBase, targetTrimmed+".exe") ||
		strings.EqualFold(cleanBase, "postcode") ||
		strings.EqualFold(cleanBase, "postcode.exe") {
		return true, true
	}

	// Fallback candidate match for prefixed names (e.g. postcode-v0.2.0-darwin-arm64)
	if strings.HasPrefix(lowerBase, "postcode") {
		return false, true
	}

	return false, false
}

// extractBinaryFromTarGz unpacks a .tar.gz archive and finds the binary payload.
func extractBinaryFromTarGz(r io.Reader, targetBinaryName string) ([]byte, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("reading gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	var fallbackBinary []byte
	var foundEmptyCandidate bool

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar entry: %w", err)
		}

		cleanName := strings.ReplaceAll(header.Name, "\\", "/")
		if header.Typeflag == tar.TypeDir || strings.HasSuffix(cleanName, "/") {
			continue
		}
		// Only extract regular files
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			continue
		}

		base := path.Base(cleanName)
		isExact, isCandidate := isExecutableCandidate(base, targetBinaryName)
		if !isCandidate {
			continue
		}

		var buf bytes.Buffer
		n, err := io.Copy(&buf, io.LimitReader(tr, maxBinaryPayloadSize+1))
		if err != nil {
			return nil, fmt.Errorf("extracting binary from tar: %w", err)
		}
		if n > maxBinaryPayloadSize {
			return nil, fmt.Errorf("binary in archive exceeds maximum size limit of %d bytes", maxBinaryPayloadSize)
		}
		if buf.Len() == 0 {
			foundEmptyCandidate = true
			continue
		}

		if isExact {
			return buf.Bytes(), nil
		}
		if len(fallbackBinary) == 0 {
			fallbackBinary = buf.Bytes()
		}
	}

	if len(fallbackBinary) > 0 {
		return fallbackBinary, nil
	}

	if foundEmptyCandidate {
		return nil, fmt.Errorf("extracted binary is empty")
	}

	return nil, fmt.Errorf("executable %q not found in release archive", targetBinaryName)
}

// extractBinaryFromZip unpacks a .zip archive and finds the binary payload.
func extractBinaryFromZip(zipBytes []byte, targetBinaryName string) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("reading zip archive: %w", err)
	}

	var fallbackBinary []byte
	var foundEmptyCandidate bool

	for _, f := range zr.File {
		cleanName := strings.ReplaceAll(f.Name, "\\", "/")
		if f.FileInfo().IsDir() || strings.HasSuffix(cleanName, "/") || !f.Mode().IsRegular() {
			continue
		}
		base := path.Base(cleanName)
		isExact, isCandidate := isExecutableCandidate(base, targetBinaryName)
		if !isCandidate {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("opening zip entry: %w", err)
		}

		var buf bytes.Buffer
		n, copyErr := io.Copy(&buf, io.LimitReader(rc, maxBinaryPayloadSize+1))
		_ = rc.Close()
		if copyErr != nil {
			return nil, fmt.Errorf("extracting binary from zip: %w", copyErr)
		}
		if n > maxBinaryPayloadSize {
			return nil, fmt.Errorf("binary in archive exceeds maximum size limit of %d bytes", maxBinaryPayloadSize)
		}
		if buf.Len() == 0 {
			foundEmptyCandidate = true
			continue
		}

		if isExact {
			return buf.Bytes(), nil
		}
		if len(fallbackBinary) == 0 {
			fallbackBinary = buf.Bytes()
		}
	}

	if len(fallbackBinary) > 0 {
		return fallbackBinary, nil
	}

	if foundEmptyCandidate {
		return nil, fmt.Errorf("extracted binary is empty")
	}

	return nil, fmt.Errorf("executable %q not found in release archive", targetBinaryName)
}

// replaceBinary safely replaces targetPath with newBinaryBytes with rollback on failure.
func replaceBinary(targetPath string, newBinaryBytes []byte) error {
	if len(newBinaryBytes) == 0 {
		return fmt.Errorf("downloaded binary payload is empty\nManual update fallback:\n    %s", fallbackInstallCmd)
	}

	// Resolve symlinks if target exists so we update the actual binary file
	if realPath, err := filepath.EvalSymlinks(targetPath); err == nil && realPath != "" {
		targetPath = realPath
	}

	var perm os.FileMode = 0755
	targetInfo, err := os.Stat(targetPath)
	if err == nil {
		if targetInfo.IsDir() {
			return fmt.Errorf("target path %s is a directory, not an executable file (run with specific binary path or install manually via '%s')", targetPath, fallbackInstallCmd)
		}
		perm = targetInfo.Mode().Perm()
		if perm&0111 == 0 {
			perm |= 0755
		}
	}

	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil && !os.IsExist(err) {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied creating directory %s (run with sudo or install manually via '%s'): %w", targetDir, fallbackInstallCmd, err)
		}
		return fmt.Errorf("creating directory %s: %w\nManual update fallback:\n    %s", targetDir, err, fallbackInstallCmd)
	}

	// Create temp file in the same directory to guarantee identical filesystem mount
	tempFile, err := os.CreateTemp(targetDir, ".postcode-update-*.tmp")
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied writing to %s (run with sudo or install manually via '%s'): %w", targetDir, fallbackInstallCmd, err)
		}
		return fmt.Errorf("creating temp file in %s: %w\nManual update fallback:\n    %s", targetDir, err, fallbackInstallCmd)
	}
	tempPath := tempFile.Name()

	cleanupTemp := true
	defer func() {
		if cleanupTemp {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.Write(newBinaryBytes); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("writing update to temp file: %w\nManual update fallback:\n    %s", err, fallbackInstallCmd)
	}
	if err := tempFile.Chmod(perm); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("setting executable permissions: %w\nManual update fallback:\n    %s", err, fallbackInstallCmd)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w\nManual update fallback:\n    %s", err, fallbackInstallCmd)
	}

	// Backup existing target binary for atomic replacement with rollback
	backupPath := targetPath + ".old"
	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		// If .old cannot be removed (e.g. locked by active process on Windows), generate unique backup
		backupPath = fmt.Sprintf("%s.old.%d", targetPath, time.Now().UnixNano())
	}

	hasBackup := false
	if _, err := os.Stat(targetPath); err == nil {
		var renameErr error
		for attempt := 0; attempt < 3; attempt++ {
			renameErr = os.Rename(targetPath, backupPath)
			if renameErr == nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		if renameErr != nil {
			if os.IsPermission(renameErr) {
				return fmt.Errorf("permission denied replacing %s (run with sudo or install manually via '%s'): %w", targetPath, fallbackInstallCmd, renameErr)
			}
			return fmt.Errorf("backing up current binary: %w\nManual update fallback:\n    %s", renameErr, fallbackInstallCmd)
		}
		hasBackup = true
	}

	// Move new binary to target path with retries for transient locks
	var renameErr error
	for attempt := 0; attempt < 3; attempt++ {
		renameErr = os.Rename(tempPath, targetPath)
		if renameErr == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if renameErr != nil {
		// Rollback on failure!
		if hasBackup {
			var rollbackErr error
			for attempt := 0; attempt < 3; attempt++ {
				rollbackErr = os.Rename(backupPath, targetPath)
				if rollbackErr == nil {
					break
				}
				time.Sleep(50 * time.Millisecond)
			}
			if rollbackErr != nil {
				return fmt.Errorf("installing new binary failed (%w); rollback also failed: %w\nManual update fallback:\n    %s", renameErr, rollbackErr, fallbackInstallCmd)
			}
			return fmt.Errorf("installing new binary (rolled back to previous version): %w\nManual update fallback:\n    %s", renameErr, fallbackInstallCmd)
		}
		return fmt.Errorf("installing new binary failed: %w\nManual update fallback:\n    %s", renameErr, fallbackInstallCmd)
	}

	// Successful update
	cleanupTemp = false

	if hasBackup {
		_ = os.Remove(backupPath) // on Windows this may fail if executable is locked; safe to ignore
	}

	return nil
}

// cleanOldBackups removes stale backup files (.old and .old.*) left by previous in-place updates.
func cleanOldBackups(basePath string) {
	if basePath == "" {
		return
	}
	_ = os.Remove(basePath + ".old")
	dir := filepath.Dir(basePath)
	base := filepath.Base(basePath)
	prefix := base + ".old."
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
				_ = os.Remove(filepath.Join(dir, e.Name()))
			}
		}
	}
}

// CompareVersions compares two version strings (semver-aware conforming to Semver 2.0.0).
// Returns:
//
//	-1 if v1 < v2 (v2 is newer)
//	 0 if v1 == v2
//	 1 if v1 > v2 (v1 is newer)
func CompareVersions(v1, v2 string) int {
	v1Clean := cleanVersionString(v1)
	v2Clean := cleanVersionString(v2)

	if v1Clean == v2Clean {
		return 0
	}

	if isDevVersion(v1Clean) && !isDevVersion(v2Clean) {
		return -1
	}
	if !isDevVersion(v1Clean) && isDevVersion(v2Clean) {
		return 1
	}
	if isDevVersion(v1Clean) && isDevVersion(v2Clean) {
		return 0
	}

	core1, pre1 := splitPreRelease(v1Clean)
	core2, pre2 := splitPreRelease(v2Clean)

	nums1 := parseVersionNumbers(core1)
	nums2 := parseVersionNumbers(core2)

	maxLen := max(len(nums1), len(nums2))
	for i := range maxLen {
		n1 := 0
		if i < len(nums1) {
			n1 = nums1[i]
		}
		n2 := 0
		if i < len(nums2) {
			n2 = nums2[i]
		}
		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	// When core numbers are equal, check pre-release tags:
	// A release version (no prerelease) is newer than a prerelease version.
	if pre1 == "" && pre2 != "" {
		return 1
	}
	if pre1 != "" && pre2 == "" {
		return -1
	}
	if pre1 != "" && pre2 != "" {
		return comparePrereleases(pre1, pre2)
	}

	return 0
}

func comparePrereleases(pre1, pre2 string) int {
	if pre1 == pre2 {
		return 0
	}
	parts1 := strings.Split(pre1, ".")
	parts2 := strings.Split(pre2, ".")
	maxLen := max(len(parts1), len(parts2))

	for i := range maxLen {
		if i >= len(parts1) {
			return -1 // pre1 has fewer identifiers, so pre2 has higher precedence
		}
		if i >= len(parts2) {
			return 1 // pre1 has more identifiers, so pre1 has higher precedence
		}
		p1, p2 := parts1[i], parts2[i]
		if p1 == p2 {
			continue
		}

		num1, err1 := strconv.Atoi(p1)
		num2, err2 := strconv.Atoi(p2)

		if err1 == nil && err2 == nil {
			// Both are numeric: compare numerically
			if num1 < num2 {
				return -1
			}
			if num1 > num2 {
				return 1
			}
		} else if err1 == nil && err2 != nil {
			// Numeric identifiers always have lower precedence than alphanumeric
			return -1
		} else if err1 != nil && err2 == nil {
			return 1
		} else {
			// If both share identical alphanumeric prefix and numeric suffix (e.g. rc2 vs rc10, beta-2 vs beta-11)
			p1Prefix, p1Num, ok1 := splitAlphaNumeric(p1)
			p2Prefix, p2Num, ok2 := splitAlphaNumeric(p2)
			if ok1 && ok2 && p1Prefix == p2Prefix {
				if p1Num < p2Num {
					return -1
				}
				if p1Num > p2Num {
					return 1
				}
				continue
			}

			// Both are alphanumeric: compare lexically
			if p1 < p2 {
				return -1
			}
			if p1 > p2 {
				return 1
			}
		}
	}
	return 0
}

func splitAlphaNumeric(s string) (prefix string, num int, ok bool) {
	if s == "" {
		return "", 0, false
	}
	i := len(s)
	for i > 0 && s[i-1] >= '0' && s[i-1] <= '9' {
		i--
	}
	if i == 0 || i == len(s) {
		return "", 0, false
	}
	val, err := strconv.Atoi(s[i:])
	if err != nil {
		return "", 0, false
	}
	return s[:i], val, true
}

func isDevVersion(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "" || v == "dev" || v == "devel" || v == "(devel)" || v == "unknown" ||
		v == "head" || v == "main" || v == "master" ||
		v == "ci-smoke" || strings.HasPrefix(v, "dev-") || strings.HasPrefix(v, "0.0.0-dev")
}

func cleanVersionString(v string) string {
	v = strings.TrimSpace(v)
	// Strip build metadata per Semver 2.0.0 (e.g. +build.1)
	if idx := strings.Index(v, "+"); idx != -1 {
		v = v[:idx]
	}
	// Strip path prefix if any (e.g. cmd/postcode/v0.1.9 or refs/tags/v0.1.9)
	if idx := strings.LastIndex(v, "/"); idx != -1 {
		v = v[idx+1:]
	}
	// Strip project name prefix if any (e.g. postcode-v0.1.9 or postcode-0.1.9)
	v = strings.TrimPrefix(v, "postcode-")
	v = strings.TrimPrefix(v, "postcode_")
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	return v
}

func splitPreRelease(v string) (string, string) {
	if idx := strings.Index(v, "-"); idx != -1 {
		return v[:idx], v[idx+1:]
	}
	return v, ""
}

func parseVersionNumbers(s string) []int {
	parts := strings.Split(s, ".")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		var n int
		_, err := fmt.Sscanf(p, "%d", &n)
		if err == nil {
			nums = append(nums, n)
		} else {
			nums = append(nums, 0)
		}
	}
	return nums
}

func executeGoInstall(ctx context.Context, out io.Writer) error {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go toolchain not found in PATH: %w\nManual update fallback:\n    %s", err, fallbackInstallCmd)
	}
	if out != nil {
		_, _ = fmt.Fprintln(out, "Executing fallback update via 'go install'...")
	}
	c := exec.CommandContext(ctx, goPath, "install", "github.com/abcubed3/postcode/cmd/postcode@latest")
	var errBuf bytes.Buffer
	if out != nil {
		c.Stdout = out
		c.Stderr = io.MultiWriter(out, &errBuf)
	} else {
		c.Stderr = &errBuf
	}
	if err := c.Run(); err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		if errMsg != "" {
			return fmt.Errorf("'go install' failed (%w): %s\nManual update fallback:\n    %s", err, errMsg, fallbackInstallCmd)
		}
		return fmt.Errorf("'go install' failed: %w\nManual update fallback:\n    %s", err, fallbackInstallCmd)
	}
	if out != nil {
		_, _ = fmt.Fprintln(out, "Successfully installed latest postcode binary via 'go install'!")
	}
	return nil
}

// NewUpdateCmd creates the update/upgrade subcommand.
func NewUpdateCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"upgrade"},
		Short:   "Check for and install updates for the postcode CLI",
		Long: `Update checks GitHub Releases for newer versions of the postcode CLI binary,
verifies cryptographic checksums, and safely replaces the executing binary in-place.`,
		Example: `  postcode update --check
  postcode update
  postcode upgrade
  postcode update -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if isOffline(v, cmd) {
				return errors.New("cannot check for updates in offline mode (network access required)")
			}

			checkOnly, _ := cmd.Flags().GetBool("check")
			force, _ := cmd.Flags().GetBool("force")
			installGo, _ := cmd.Flags().GetBool("install-go")

			githubAPI, _ := cmd.Flags().GetString("github-api")
			if githubAPI == "" {
				githubAPI = v.GetString("github-api-url")
			}
			if githubAPI == "" {
				githubAPI = defaultGitHubAPIURL
			}

			githubToken, _ := cmd.Flags().GetString("github-token")
			if githubToken == "" {
				githubToken = v.GetString("github-token")
			}

			targetPath, _ := cmd.Flags().GetString("target-path")
			currentVer, _ := cmd.Flags().GetString("current-version")
			if currentVer == "" {
				currentVer = Version
			}

			timeout := v.GetDuration("timeout")
			if timeout <= 0 {
				timeout = 30 * time.Second
			}

			client := &http.Client{Timeout: timeout}
			if defaultHTTPTransport != nil {
				client.Transport = defaultHTTPTransport
			}

			if targetPath != "" {
				cleanOldBackups(targetPath)
			} else if execPath, err := os.Executable(); err == nil {
				cleanOldBackups(execPath)
				if realPath, err := filepath.EvalSymlinks(execPath); err == nil && realPath != execPath {
					cleanOldBackups(realPath)
				}
			}

			updater := NewUpdater(UpdaterConfig{
				CurrentVersion:   currentVer,
				GitHubAPIURL:     githubAPI,
				GitHubToken:      githubToken,
				TargetExecutable: targetPath,
				HTTPClient:       client,
			})

			checkResult, release, err := updater.Check(cmd.Context())
			if err != nil {
				if installGo && !checkOnly {
					return executeGoInstall(cmd.Context(), cmd.OutOrStdout())
				}
				return err
			}

			isStructured := strings.ToLower(v.GetString("output")) != "" && strings.ToLower(v.GetString("output")) != string(FormatText)

			if checkOnly {
				return PrintOutput(cmd, v, checkResult, func(w io.Writer) error {
					if checkResult.UpdateAvailable {
						_, _ = fmt.Fprintln(w, "Update available!")
						_, _ = fmt.Fprintf(w, "  Current version: %s\n", checkResult.CurrentVersion)
						_, _ = fmt.Fprintf(w, "  Latest version:  %s\n", checkResult.LatestVersion)
						if checkResult.ReleaseURL != "" {
							_, _ = fmt.Fprintf(w, "  Release notes:   %s\n", checkResult.ReleaseURL)
						}
						_, _ = fmt.Fprintln(w, "\nRun 'postcode update' to install the update.")
					} else {
						_, _ = fmt.Fprintf(w, "postcode is already up-to-date (%s).\n", checkResult.CurrentVersion)
					}
					return nil
				})
			}

			if !checkResult.UpdateAvailable && !force {
				return PrintOutput(cmd, v, checkResult, func(w io.Writer) error {
					_, _ = fmt.Fprintf(w, "postcode is already up-to-date (%s).\n", checkResult.CurrentVersion)
					return nil
				})
			}

			var progressWriter io.Writer
			if !isStructured {
				progressWriter = cmd.OutOrStdout()
				_, _ = fmt.Fprintf(progressWriter, "Found newer version: %s (current: %s)\n", checkResult.LatestVersion, checkResult.CurrentVersion)
			}

			upgradeResult, err := updater.Upgrade(cmd.Context(), release, progressWriter)
			if err != nil {
				if installGo {
					return executeGoInstall(cmd.Context(), cmd.OutOrStdout())
				}
				return err
			}

			return PrintOutput(cmd, v, upgradeResult, func(w io.Writer) error {
				_, _ = fmt.Fprintf(w, "Successfully updated postcode to %s!\n", upgradeResult.UpdatedVersion)
				return nil
			})
		},
	}

	cmd.Flags().BoolP("check", "c", false, "check for updates without modifying the local binary")
	cmd.Flags().BoolP("force", "f", false, "force update even if already on the latest version")
	cmd.Flags().Bool("install-go", false, "fallback to running 'go install' if binary update fails or is unsupported")
	cmd.Flags().String("github-token", "", "GitHub personal access token for authenticated API requests")

	// Testing & Advanced flags (hidden from general help)
	cmd.Flags().String("github-api", "", "GitHub API base URL override (internal/testing)")
	cmd.Flags().String("target-path", "", "Target executable path override (internal/testing)")
	cmd.Flags().String("current-version", "", "Current version override (internal/testing)")
	_ = cmd.Flags().MarkHidden("github-api")
	_ = cmd.Flags().MarkHidden("target-path")
	_ = cmd.Flags().MarkHidden("current-version")

	return cmd
}
