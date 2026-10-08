package commands_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	cmd "github.com/abcubed3/postcode/cmd/postcode/internal/commands"
)

type mockTransport struct {
	handler http.Handler
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	m.handler.ServeHTTP(rec, req)
	resp := rec.Result()
	resp.Request = req
	return resp, nil
}

func registerMockServer(t *testing.T, handler http.Handler) string {
	t.Helper()
	mt := &mockTransport{handler: handler}
	cmd.SetDefaultHTTPTransport(mt)
	t.Cleanup(func() {
		cmd.SetDefaultHTTPTransport(nil)
	})
	return "http://mock-github-api.local"
}

func createTestTarGz(filename string, content []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: filename,
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}
	if _, err := tw.Write(content); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func createTestZip(filename string, content []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	f, err := zw.Create(filename)
	if err != nil {
		return nil, err
	}
	if _, err := f.Write(content); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func computeSHA256(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int // -1 (v1 < v2), 0 (v1 == v2), 1 (v1 > v2)
	}{
		{"v0.1.0", "v0.2.0", -1},
		{"v0.2.0", "v0.2.0", 0},
		{"v0.2.1", "v0.2.0", 1},
		{"0.1.7", "v0.1.7", 0},
		{"v1.0.0", "1.0.0", 0},
		{"cmd/postcode/v0.2.0", "v0.2.0", 0},
		{"v1.9.0", "v1.10.0", -1},
		{"v1.10.0", "v1.9.0", 1},
		{"v1.0", "v1.0.0", 0},
		{"v1.0.0-rc1", "v1.0.0", -1},
		{"v1.0.0", "v1.0.0-rc1", 1},
		{"v1.0.0-beta", "v1.0.0-alpha", 1},
		{"v1.0.0-beta.2", "v1.0.0-beta.11", -1},
		{"v1.0.0-beta.11", "v1.0.0-beta.2", 1},
		{"v1.0.0-rc.1", "v1.0.0-rc.2", -1},
		{"v1.0.0-alpha", "v1.0.0-alpha.1", -1},
		{"v1.0.0-alpha.1", "v1.0.0-alpha.beta", -1},
		{"1.0.0+20130313144700", "1.0.0", 0},
		{"postcode-v0.2.0", "v0.2.0", 0},
		{"postcode-0.2.0", "v0.1.9", 1},
		{"dev", "v0.1.0", -1},
		{"dev-12345", "v0.1.0", -1},
		{"(devel)", "v0.1.0", -1},
		{"ci-smoke", "v0.1.0", -1},
		{"v0.1.0", "dev", 1},
		{"v1.0.0-rc2", "v1.0.0-rc10", -1},
		{"v1.0.0-rc10", "v1.0.0-rc2", 1},
		{"v1.0.0-beta-2", "v1.0.0-beta-11", -1},
		{"v1.0.0-patch1", "v1.0.0-patch2", -1},
		{"HEAD", "v0.1.0", -1},
		{"main", "v0.1.0", -1},
		{"master", "v0.1.0", -1},
		{"dev", "dev", 0},
		{"unknown", "v1.0.0", -1},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.v1, tt.v2), func(t *testing.T) {
			got := cmd.CompareVersions(tt.v1, tt.v2)
			if got != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.expected)
			}
		})
	}
}

func TestCLI_UpdateCheck(t *testing.T) {
	releaseJSON := cmd.GitHubRelease{
		TagName:     "v0.2.0",
		Name:        "Postcode v0.2.0",
		PublishedAt: "2026-10-07T12:00:00Z",
		HTMLURL:     "https://github.com/abcubed3/postcode/releases/tag/v0.2.0",
	}

	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(releaseJSON)
	}))

	t.Run("check when update is available", func(t *testing.T) {
		out, _, err := executeCmd([]string{
			"update",
			"--check",
			"--github-api", mockURL,
			"--current-version", "v0.1.0",
		}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Update available!") {
			t.Errorf("expected 'Update available!' in output, got: %s", out)
		}
		if !strings.Contains(out, "Latest version:  v0.2.0") {
			t.Errorf("expected 'Latest version:  v0.2.0' in output, got: %s", out)
		}
		if !strings.Contains(out, "Current version: v0.1.0") {
			t.Errorf("expected 'Current version: v0.1.0' in output, got: %s", out)
		}
	})

	t.Run("check when already up-to-date", func(t *testing.T) {
		out, _, err := executeCmd([]string{
			"update",
			"--check",
			"--github-api", mockURL,
			"--current-version", "v0.2.0",
		}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "postcode is already up-to-date (v0.2.0)") {
			t.Errorf("expected up-to-date message in output, got: %s", out)
		}
	})

	t.Run("check when running dev version", func(t *testing.T) {
		out, _, err := executeCmd([]string{
			"update",
			"--check",
			"--github-api", mockURL,
			"--current-version", "dev",
		}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Update available!") {
			t.Errorf("expected update available for dev version, got: %s", out)
		}
	})

	t.Run("check json format output", func(t *testing.T) {
		out, _, err := executeCmd([]string{
			"update",
			"--check",
			"--github-api", mockURL,
			"--current-version", "v0.1.0",
			"-o", "json",
		}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var res cmd.UpdateCheckResult
		if err := json.Unmarshal([]byte(out), &res); err != nil {
			t.Fatalf("failed to decode JSON output: %v\nOutput: %s", err, out)
		}
		if !res.UpdateAvailable {
			t.Error("expected update_available to be true")
		}
		if res.LatestVersion != "v0.2.0" {
			t.Errorf("expected latest_version v0.2.0, got %s", res.LatestVersion)
		}
		if res.CurrentVersion != "v0.1.0" {
			t.Errorf("expected current_version v0.1.0, got %s", res.CurrentVersion)
		}
	})

	t.Run("upgrade alias works identically", func(t *testing.T) {
		out, _, err := executeCmd([]string{
			"upgrade",
			"--check",
			"--github-api", mockURL,
			"--current-version", "v0.1.0",
		}, "")
		if err != nil {
			t.Fatalf("unexpected error with upgrade alias: %v", err)
		}
		if !strings.Contains(out, "Update available!") {
			t.Errorf("expected 'Update available!' in upgrade output, got: %s", out)
		}
	})
}

func TestCLI_Update_AlreadyUpToDate(t *testing.T) {
	releaseJSON := cmd.GitHubRelease{
		TagName:     "v0.2.0",
		Name:        "Postcode v0.2.0",
		PublishedAt: "2026-10-07T12:00:00Z",
		HTMLURL:     "https://github.com/abcubed3/postcode/releases/tag/v0.2.0",
	}

	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(releaseJSON)
	}))

	out, _, err := executeCmd([]string{
		"update",
		"--github-api", mockURL,
		"--current-version", "v0.2.0",
	}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "postcode is already up-to-date (v0.2.0)") {
		t.Errorf("expected up-to-date message, got: %s", out)
	}
}

func TestCLI_Update_SuccessfulUpgrade(t *testing.T) {
	newBinaryContent := []byte("#!/bin/sh\necho 'postcode v0.2.0 simulated binary'\n")

	// Create archive matching host OS and Arch
	var archiveBytes []byte
	var archiveName string
	var err error

	if runtime.GOOS == "windows" {
		archiveName = fmt.Sprintf("postcode-windows-%s.zip", runtime.GOARCH)
		archiveBytes, err = createTestZip(fmt.Sprintf("postcode-windows-%s.exe", runtime.GOARCH), newBinaryContent)
	} else {
		archiveName = fmt.Sprintf("postcode-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		archiveBytes, err = createTestTarGz(fmt.Sprintf("postcode-%s-%s", runtime.GOOS, runtime.GOARCH), newBinaryContent)
	}
	if err != nil {
		t.Fatalf("failed to create test archive: %v", err)
	}

	checksum := computeSHA256(archiveBytes)
	checksumsContent := fmt.Sprintf("%s  %s\n", checksum, archiveName)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName:     "v0.2.0",
				Name:        "Postcode v0.2.0",
				PublishedAt: "2026-10-07T12:00:00Z",
				HTMLURL:     "https://github.com/abcubed3/postcode/releases/tag/v0.2.0",
				Assets: []cmd.GitHubAsset{
					{
						Name:               archiveName,
						BrowserDownloadURL: mockBase + "/download/" + archiveName,
						Size:               int64(len(archiveBytes)),
					},
					{
						Name:               "checksums.txt",
						BrowserDownloadURL: mockBase + "/download/checksums.txt",
						Size:               int64(len(checksumsContent)),
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/" + archiveName:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(archiveBytes)
		case "/download/checksums.txt":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(checksumsContent))
		default:
			http.NotFound(w, r)
		}
	}))

	// Prepare dummy target binary
	tempDir := t.TempDir()
	targetBinary := filepath.Join(tempDir, "postcode-binary")
	originalContent := []byte("#!/bin/sh\necho 'old binary'\n")
	if err := os.WriteFile(targetBinary, originalContent, 0755); err != nil {
		t.Fatalf("failed to write original binary: %v", err)
	}

	out, _, err := executeCmd([]string{
		"update",
		"--github-api", mockURL,
		"--current-version", "v0.1.0",
		"--target-path", targetBinary,
	}, "")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if !strings.Contains(out, "Successfully updated postcode to v0.2.0") {
		t.Errorf("expected success message, got: %s", out)
	}

	// Verify target binary was replaced with new content
	updatedContent, err := os.ReadFile(targetBinary)
	if err != nil {
		t.Fatalf("failed to read target binary after update: %v", err)
	}
	if !bytes.Equal(updatedContent, newBinaryContent) {
		t.Errorf("target binary content mismatch: got %q, want %q", updatedContent, newBinaryContent)
	}
}

func TestCLI_Update_ChecksumMismatch(t *testing.T) {
	corruptedArchive := []byte("corrupted payload")
	var archiveName string
	if runtime.GOOS == "windows" {
		archiveName = fmt.Sprintf("postcode-windows-%s.zip", runtime.GOARCH)
	} else {
		archiveName = fmt.Sprintf("postcode-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	}

	// Deliberately wrong checksum
	wrongChecksum := "0000000000000000000000000000000000000000000000000000000000000000"
	checksumsContent := fmt.Sprintf("%s  %s\n", wrongChecksum, archiveName)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.2.0",
				Assets: []cmd.GitHubAsset{
					{
						Name:               archiveName,
						BrowserDownloadURL: mockBase + "/download/" + archiveName,
					},
					{
						Name:               "checksums.txt",
						BrowserDownloadURL: mockBase + "/download/checksums.txt",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/" + archiveName:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(corruptedArchive)
		case "/download/checksums.txt":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(checksumsContent))
		default:
			http.NotFound(w, r)
		}
	}))

	tempDir := t.TempDir()
	targetBinary := filepath.Join(tempDir, "postcode-target")
	originalContent := []byte("original uncorrupted binary")
	if err := os.WriteFile(targetBinary, originalContent, 0755); err != nil {
		t.Fatalf("failed to create target binary: %v", err)
	}

	_, _, err := executeCmd([]string{
		"update",
		"--github-api", mockURL,
		"--current-version", "v0.1.0",
		"--target-path", targetBinary,
	}, "")

	if err == nil {
		t.Fatal("expected error on checksum mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "checksum verification failed") {
		t.Errorf("expected checksum verification error, got: %v", err)
	}

	// Verify original file is unchanged
	content, err := os.ReadFile(targetBinary)
	if err != nil {
		t.Fatalf("failed to read target binary: %v", err)
	}
	if !bytes.Equal(content, originalContent) {
		t.Errorf("target binary was modified despite checksum mismatch: got %q, want %q", content, originalContent)
	}
}

func TestCLI_Update_MissingAsset_Fallback(t *testing.T) {
	// Release only has assets for a different OS/arch (e.g. wasm)
	releaseJSON := cmd.GitHubRelease{
		TagName: "v0.2.0",
		Assets: []cmd.GitHubAsset{
			{
				Name: "postcode-wasm.tar.gz",
			},
			{
				Name: "checksums.txt",
			},
		},
	}

	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(releaseJSON)
	}))

	_, _, err := executeCmd([]string{
		"update",
		"--github-api", mockURL,
		"--current-version", "v0.1.0",
	}, "")

	if err == nil {
		t.Fatal("expected error for missing platform asset, got nil")
	}
	if !strings.Contains(err.Error(), "go install github.com/abcubed3/postcode/cmd/postcode@latest") {
		t.Errorf("expected fallback install command in error, got: %v", err)
	}
}

func TestCLI_Update_OfflineMode(t *testing.T) {
	_, _, err := executeCmd([]string{"update", "--offline"}, "")
	if err == nil {
		t.Fatal("expected error when update is run in offline mode, got nil")
	}
	if !strings.Contains(err.Error(), "offline mode") {
		t.Errorf("expected offline error message, got: %v", err)
	}
}

func TestCLI_Update_Force(t *testing.T) {
	newBinaryContent := []byte("forced update binary\n")
	var archiveBytes []byte
	var archiveName string
	var err error

	if runtime.GOOS == "windows" {
		archiveName = fmt.Sprintf("postcode-windows-%s.zip", runtime.GOARCH)
		archiveBytes, err = createTestZip(fmt.Sprintf("postcode-windows-%s.exe", runtime.GOARCH), newBinaryContent)
	} else {
		archiveName = fmt.Sprintf("postcode-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		archiveBytes, err = createTestTarGz(fmt.Sprintf("postcode-%s-%s", runtime.GOOS, runtime.GOARCH), newBinaryContent)
	}
	if err != nil {
		t.Fatalf("failed to create test archive: %v", err)
	}

	checksum := computeSHA256(archiveBytes)
	checksumsContent := fmt.Sprintf("%s  %s\n", checksum, archiveName)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.2.0",
				Assets: []cmd.GitHubAsset{
					{
						Name:               archiveName,
						BrowserDownloadURL: mockBase + "/download/" + archiveName,
					},
					{
						Name:               "checksums.txt",
						BrowserDownloadURL: mockBase + "/download/checksums.txt",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/" + archiveName:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(archiveBytes)
		case "/download/checksums.txt":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(checksumsContent))
		default:
			http.NotFound(w, r)
		}
	}))

	tempDir := t.TempDir()
	targetBinary := filepath.Join(tempDir, "postcode-target")
	if err := os.WriteFile(targetBinary, []byte("existing"), 0755); err != nil {
		t.Fatalf("failed to write target binary: %v", err)
	}

	// Current version is already v0.2.0, but --force flag is passed
	out, _, err := executeCmd([]string{
		"update",
		"--force",
		"--github-api", mockURL,
		"--current-version", "v0.2.0",
		"--target-path", targetBinary,
	}, "")
	if err != nil {
		t.Fatalf("force update failed: %v", err)
	}
	if !strings.Contains(out, "Successfully updated") {
		t.Errorf("expected success message with --force, got: %s", out)
	}

	updated, err := os.ReadFile(targetBinary)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}
	if !bytes.Equal(updated, newBinaryContent) {
		t.Errorf("expected binary content %q, got %q", newBinaryContent, updated)
	}
}

func TestCLI_Update_CrossPlatformArchiveExtraction(t *testing.T) {
	// Test extracting tar.gz (Unix)
	tarContent := []byte("linux-arm64-binary-data")
	tarBytes, err := createTestTarGz("postcode-linux-arm64", tarContent)
	if err != nil {
		t.Fatalf("createTestTarGz failed: %v", err)
	}

	// Test extracting zip (Windows)
	zipContent := []byte("windows-amd64-binary-data")
	zipBytes, err := createTestZip("postcode-windows-amd64.exe", zipContent)
	if err != nil {
		t.Fatalf("createTestZip failed: %v", err)
	}

	tarChecksum := computeSHA256(tarBytes)
	zipChecksum := computeSHA256(zipBytes)

	checksums := fmt.Sprintf("%s  postcode-linux-arm64.tar.gz\n%s  postcode-windows-amd64.zip\n", tarChecksum, zipChecksum)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.3.0",
				Assets: []cmd.GitHubAsset{
					{Name: "postcode-linux-arm64.tar.gz", BrowserDownloadURL: mockBase + "/download/linux"},
					{Name: "postcode-windows-amd64.zip", BrowserDownloadURL: mockBase + "/download/windows"},
					{Name: "checksums.txt", BrowserDownloadURL: mockBase + "/download/checksums.txt"},
				},
			}
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/linux":
			_, _ = w.Write(tarBytes)
		case "/download/windows":
			_, _ = w.Write(zipBytes)
		case "/download/checksums.txt":
			_, _ = w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	}))

	t.Run("linux arm64 upgrade flow via Updater", func(t *testing.T) {
		tempDir := t.TempDir()
		target := filepath.Join(tempDir, "postcode")
		_ = os.WriteFile(target, []byte("old"), 0755)

		updater := cmd.NewUpdater(cmd.UpdaterConfig{
			CurrentVersion:   "v0.1.0",
			GitHubAPIURL:     mockURL,
			TargetExecutable: target,
			GOOS:             "linux",
			GOARCH:           "arm64",
		})

		res, err := updater.Upgrade(t.Context(), nil, nil)
		if err != nil {
			t.Fatalf("linux upgrade failed: %v", err)
		}
		if res.UpdatedVersion != "v0.3.0" {
			t.Errorf("got %s, want v0.3.0", res.UpdatedVersion)
		}

		data, _ := os.ReadFile(target)
		if !bytes.Equal(data, tarContent) {
			t.Errorf("expected extracted linux content %q, got %q", tarContent, data)
		}
	})

	t.Run("windows amd64 upgrade flow via Updater", func(t *testing.T) {
		tempDir := t.TempDir()
		target := filepath.Join(tempDir, "postcode.exe")
		_ = os.WriteFile(target, []byte("old"), 0755)

		updater := cmd.NewUpdater(cmd.UpdaterConfig{
			CurrentVersion:   "v0.1.0",
			GitHubAPIURL:     mockURL,
			TargetExecutable: target,
			GOOS:             "windows",
			GOARCH:           "amd64",
		})

		res, err := updater.Upgrade(t.Context(), nil, nil)
		if err != nil {
			t.Fatalf("windows upgrade failed: %v", err)
		}
		if res.UpdatedVersion != "v0.3.0" {
			t.Errorf("got %s, want v0.3.0", res.UpdatedVersion)
		}

		data, _ := os.ReadFile(target)
		if !bytes.Equal(data, zipContent) {
			t.Errorf("expected extracted windows content %q, got %q", zipContent, data)
		}
	})
}

func TestCLI_Update_BSDChecksumFormat(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho 'bsd-checksum-test'\n")
	archiveBytes, err := createTestTarGz("postcode-darwin-arm64", binaryContent)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	checksum := computeSHA256(archiveBytes)
	// BSD format: SHA256 (filename) = hash
	bsdChecksums := fmt.Sprintf("SHA256 (postcode-darwin-arm64.tar.gz) = %s\n", checksum)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.4.0",
				Assets: []cmd.GitHubAsset{
					{Name: "postcode-darwin-arm64.tar.gz", BrowserDownloadURL: mockBase + "/download/archive.tar.gz"},
					{Name: "checksums.txt", BrowserDownloadURL: mockBase + "/download/checksums.txt"},
				},
			}
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/archive.tar.gz":
			_, _ = w.Write(archiveBytes)
		case "/download/checksums.txt":
			_, _ = w.Write([]byte(bsdChecksums))
		default:
			http.NotFound(w, r)
		}
	}))

	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "postcode")
	_ = os.WriteFile(target, []byte("old"), 0755)

	updater := cmd.NewUpdater(cmd.UpdaterConfig{
		CurrentVersion:   "v0.1.0",
		GitHubAPIURL:     mockURL,
		TargetExecutable: target,
		GOOS:             "darwin",
		GOARCH:           "arm64",
	})

	res, err := updater.Upgrade(t.Context(), nil, nil)
	if err != nil {
		t.Fatalf("upgrade with BSD checksum format failed: %v", err)
	}
	if res.UpdatedVersion != "v0.4.0" {
		t.Errorf("got %s, want v0.4.0", res.UpdatedVersion)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}
	if !bytes.Equal(data, binaryContent) {
		t.Errorf("target content mismatch: got %q, want %q", data, binaryContent)
	}
}

func TestCLI_Update_ArchiveWithAuxiliaryFiles(t *testing.T) {
	// Create tar archive with non-binary auxiliary files listed BEFORE the actual executable
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	realBinaryContent := []byte("#!/bin/sh\necho 'real-binary'\n")

	files := []struct {
		name    string
		mode    int64
		content []byte
	}{
		{"postcode.yaml", 0644, []byte("api_key: test\n")},
		{"postcode.1", 0644, []byte(".TH POSTCODE 1\n")},
		{"postcode-completion.bash", 0644, []byte("complete -F _postcode postcode\n")},
		{"postcode-linux-amd64", 0755, realBinaryContent},
	}

	for _, f := range files {
		hdr := &tar.Header{
			Name: f.name,
			Mode: f.mode,
			Size: int64(len(f.content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("failed to write header for %s: %v", f.name, err)
		}
		if _, err := tw.Write(f.content); err != nil {
			t.Fatalf("failed to write content for %s: %v", f.name, err)
		}
	}
	_ = tw.Close()
	_ = gw.Close()

	archiveBytes := buf.Bytes()
	checksum := computeSHA256(archiveBytes)
	checksums := fmt.Sprintf("%s  postcode-linux-amd64.tar.gz\n", checksum)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.5.0",
				Assets: []cmd.GitHubAsset{
					{Name: "postcode-linux-amd64.tar.gz", BrowserDownloadURL: mockBase + "/download/archive.tar.gz"},
					{Name: "checksums.txt", BrowserDownloadURL: mockBase + "/download/checksums.txt"},
				},
			}
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/archive.tar.gz":
			_, _ = w.Write(archiveBytes)
		case "/download/checksums.txt":
			_, _ = w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	}))

	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "postcode")
	_ = os.WriteFile(target, []byte("old"), 0755)

	updater := cmd.NewUpdater(cmd.UpdaterConfig{
		CurrentVersion:   "v0.1.0",
		GitHubAPIURL:     mockURL,
		TargetExecutable: target,
		GOOS:             "linux",
		GOARCH:           "amd64",
	})

	_, err := updater.Upgrade(t.Context(), nil, nil)
	if err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read updated binary: %v", err)
	}
	if !bytes.Equal(data, realBinaryContent) {
		t.Fatalf("auxiliary file was mistakenly extracted instead of real binary! got: %q", string(data))
	}
}

func TestCLI_Update_TargetIsDirectory(t *testing.T) {
	binaryContent := []byte("binary")
	archiveBytes, err := createTestTarGz("postcode-darwin-arm64", binaryContent)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	checksum := computeSHA256(archiveBytes)
	checksums := fmt.Sprintf("%s  postcode-darwin-arm64.tar.gz\n", checksum)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.6.0",
				Assets: []cmd.GitHubAsset{
					{Name: "postcode-darwin-arm64.tar.gz", BrowserDownloadURL: mockBase + "/download/archive.tar.gz"},
					{Name: "checksums.txt", BrowserDownloadURL: mockBase + "/download/checksums.txt"},
				},
			}
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/archive.tar.gz":
			_, _ = w.Write(archiveBytes)
		case "/download/checksums.txt":
			_, _ = w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	}))

	dirTarget := t.TempDir() // Target is a directory, not a file!
	updater := cmd.NewUpdater(cmd.UpdaterConfig{
		CurrentVersion:   "v0.1.0",
		GitHubAPIURL:     mockURL,
		TargetExecutable: dirTarget,
		GOOS:             "darwin",
		GOARCH:           "arm64",
	})

	_, err = updater.Upgrade(t.Context(), nil, nil)
	if err == nil {
		t.Fatal("expected error when target path is a directory, got nil")
	}
	if !strings.Contains(err.Error(), "is a directory, not an executable file") {
		t.Errorf("expected directory rejection error, got: %v", err)
	}
}

func TestCLI_Update_WriteProtectedDirectory_Fallback(t *testing.T) {
	binaryContent := []byte("binary")
	archiveBytes, err := createTestTarGz("postcode-darwin-arm64", binaryContent)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	checksum := computeSHA256(archiveBytes)
	checksums := fmt.Sprintf("%s  postcode-darwin-arm64.tar.gz\n", checksum)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.7.0",
				Assets: []cmd.GitHubAsset{
					{Name: "postcode-darwin-arm64.tar.gz", BrowserDownloadURL: mockBase + "/download/archive.tar.gz"},
					{Name: "checksums.txt", BrowserDownloadURL: mockBase + "/download/checksums.txt"},
				},
			}
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/archive.tar.gz":
			_, _ = w.Write(archiveBytes)
		case "/download/checksums.txt":
			_, _ = w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	}))

	baseDir := t.TempDir()
	readOnlyDir := filepath.Join(baseDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("failed to create readonly dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(readOnlyDir, 0755)
	})

	target := filepath.Join(readOnlyDir, "postcode")
	updater := cmd.NewUpdater(cmd.UpdaterConfig{
		CurrentVersion:   "v0.1.0",
		GitHubAPIURL:     mockURL,
		TargetExecutable: target,
		GOOS:             "darwin",
		GOARCH:           "arm64",
	})

	_, err = updater.Upgrade(t.Context(), nil, nil)
	if err == nil {
		// If running as root in container/environment, skip permission check
		t.Skip("skipping write-protected test: process has root permissions")
	}

	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected permission denied error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "go install github.com/abcubed3/postcode/cmd/postcode@latest") {
		t.Errorf("expected go install fallback instructions in error, got: %v", err)
	}
}

func TestCLI_Update_StaleBackupCleanup(t *testing.T) {
	binaryContent := []byte("new-binary")
	var archiveBytes []byte
	var archiveName string
	var err error

	if runtime.GOOS == "windows" {
		archiveName = fmt.Sprintf("postcode-windows-%s.zip", runtime.GOARCH)
		archiveBytes, err = createTestZip(fmt.Sprintf("postcode-windows-%s.exe", runtime.GOARCH), binaryContent)
	} else {
		archiveName = fmt.Sprintf("postcode-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		archiveBytes, err = createTestTarGz(fmt.Sprintf("postcode-%s-%s", runtime.GOOS, runtime.GOARCH), binaryContent)
	}
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	checksum := computeSHA256(archiveBytes)
	checksums := fmt.Sprintf("%s  %s\n", checksum, archiveName)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.8.0",
				Assets: []cmd.GitHubAsset{
					{Name: archiveName, BrowserDownloadURL: mockBase + "/download/" + archiveName},
					{Name: "checksums.txt", BrowserDownloadURL: mockBase + "/download/checksums.txt"},
				},
			}
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/" + archiveName:
			_, _ = w.Write(archiveBytes)
		case "/download/checksums.txt":
			_, _ = w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	}))

	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "postcode")
	staleBackup := target + ".old"
	_ = os.WriteFile(target, []byte("old-binary"), 0755)
	_ = os.WriteFile(staleBackup, []byte("stale-backup-from-previous-run"), 0755)

	out, _, err := executeCmd([]string{
		"update",
		"--github-api", mockURL,
		"--current-version", "v0.1.0",
		"--target-path", target,
	}, "")
	if err != nil {
		t.Fatalf("update failed with stale backup present: %v", err)
	}
	if !strings.Contains(out, "Successfully updated") {
		t.Errorf("expected success message, got: %s", out)
	}

	if _, err := os.Stat(staleBackup); !os.IsNotExist(err) {
		t.Errorf("stale backup %s was not cleaned up after update", staleBackup)
	}
}

func TestCLI_Update_ReleaseListFallback(t *testing.T) {
	rel := cmd.GitHubRelease{
		TagName:     "v0.9.0",
		Name:        "Postcode v0.9.0",
		PublishedAt: "2026-10-07T12:00:00Z",
		HTMLURL:     "https://github.com/abcubed3/postcode/releases/tag/v0.9.0",
	}

	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			// Simulate /releases/latest returning 404 (e.g. only prereleases or no latest tag)
			http.NotFound(w, r)
		case "/repos/abcubed3/postcode/releases":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]cmd.GitHubRelease{rel})
		default:
			http.NotFound(w, r)
		}
	}))

	out, _, err := executeCmd([]string{
		"update",
		"--check",
		"--github-api", mockURL,
		"--current-version", "v0.1.0",
	}, "")
	if err != nil {
		t.Fatalf("expected fallback to /releases list to succeed: %v", err)
	}
	if !strings.Contains(out, "Latest version:  v0.9.0") {
		t.Errorf("expected latest version v0.9.0 from fallback list, got: %s", out)
	}
}

func TestCLI_Update_EmptyBinaryRejection(t *testing.T) {
	var archiveBytes []byte
	var archiveName string
	var err error

	if runtime.GOOS == "windows" {
		archiveName = fmt.Sprintf("postcode-windows-%s.zip", runtime.GOARCH)
		archiveBytes, err = createTestZip(fmt.Sprintf("postcode-windows-%s.exe", runtime.GOARCH), []byte{})
	} else {
		archiveName = fmt.Sprintf("postcode-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		archiveBytes, err = createTestTarGz(fmt.Sprintf("postcode-%s-%s", runtime.GOOS, runtime.GOARCH), []byte{})
	}
	if err != nil {
		t.Fatalf("failed to create test archive: %v", err)
	}

	checksum := computeSHA256(archiveBytes)
	checksumsContent := fmt.Sprintf("%s  %s\n", checksum, archiveName)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.9.1",
				Assets: []cmd.GitHubAsset{
					{Name: archiveName, BrowserDownloadURL: mockBase + "/download/" + archiveName},
					{Name: "checksums.txt", BrowserDownloadURL: mockBase + "/download/checksums.txt"},
				},
			}
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/" + archiveName:
			_, _ = w.Write(archiveBytes)
		case "/download/checksums.txt":
			_, _ = w.Write([]byte(checksumsContent))
		default:
			http.NotFound(w, r)
		}
	}))

	tempDir := t.TempDir()
	targetBinary := filepath.Join(tempDir, "postcode-target")
	origContent := []byte("original binary")
	_ = os.WriteFile(targetBinary, origContent, 0755)

	_, _, err = executeCmd([]string{
		"update",
		"--github-api", mockURL,
		"--current-version", "v0.1.0",
		"--target-path", targetBinary,
	}, "")
	if err == nil {
		t.Fatal("expected error on empty binary payload, got nil")
	}
	if !strings.Contains(err.Error(), "extracted binary is empty") {
		t.Errorf("expected 'extracted binary is empty' error, got: %v", err)
	}

	// Verify original binary is unmodified
	data, _ := os.ReadFile(targetBinary)
	if !bytes.Equal(data, origContent) {
		t.Errorf("original binary was overwritten with empty file!")
	}
}

func TestCLI_Update_SymlinkTargetPath(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho 'symlink-test'\n")
	var archiveBytes []byte
	var archiveName string
	var err error

	if runtime.GOOS == "windows" {
		archiveName = fmt.Sprintf("postcode-windows-%s.zip", runtime.GOARCH)
		archiveBytes, err = createTestZip(fmt.Sprintf("postcode-windows-%s.exe", runtime.GOARCH), binaryContent)
	} else {
		archiveName = fmt.Sprintf("postcode-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		archiveBytes, err = createTestTarGz(fmt.Sprintf("postcode-%s-%s", runtime.GOOS, runtime.GOARCH), binaryContent)
	}
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	checksum := computeSHA256(archiveBytes)
	checksumsContent := fmt.Sprintf("%s  %s\n", checksum, archiveName)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.9.2",
				Assets: []cmd.GitHubAsset{
					{Name: archiveName, BrowserDownloadURL: mockBase + "/download/" + archiveName},
					{Name: "checksums.txt", BrowserDownloadURL: mockBase + "/download/checksums.txt"},
				},
			}
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/" + archiveName:
			_, _ = w.Write(archiveBytes)
		case "/download/checksums.txt":
			_, _ = w.Write([]byte(checksumsContent))
		default:
			http.NotFound(w, r)
		}
	}))

	tempDir := t.TempDir()
	realBinDir := filepath.Join(tempDir, "real")
	_ = os.MkdirAll(realBinDir, 0755)
	realBinary := filepath.Join(realBinDir, "postcode-real")
	_ = os.WriteFile(realBinary, []byte("old-real"), 0755)

	symlinkPath := filepath.Join(tempDir, "postcode-symlink")
	if err := os.Symlink(realBinary, symlinkPath); err != nil {
		t.Skip("skipping symlink test: symlinks not supported in environment")
	}

	_, _, err = executeCmd([]string{
		"update",
		"--github-api", mockURL,
		"--current-version", "v0.1.0",
		"--target-path", symlinkPath,
	}, "")
	if err != nil {
		t.Fatalf("update with symlink target failed: %v", err)
	}

	// Verify symlink is still a symlink
	lstat, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("lstat symlink failed: %v", err)
	}
	if lstat.Mode()&os.ModeSymlink == 0 {
		t.Errorf("target was converted from symlink to regular file!")
	}

	// Verify underlying real file was updated
	updatedReal, err := os.ReadFile(realBinary)
	if err != nil {
		t.Fatalf("reading real binary failed: %v", err)
	}
	if !bytes.Equal(updatedReal, binaryContent) {
		t.Errorf("real binary was not updated: got %q, want %q", updatedReal, binaryContent)
	}
}

func TestCLI_Update_CheckWithInstallGo_NonDestructive(t *testing.T) {
	// API returns 500 error
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))

	tempDir := t.TempDir()
	targetBinary := filepath.Join(tempDir, "postcode")
	origContent := []byte("orig-binary")
	_ = os.WriteFile(targetBinary, origContent, 0755)

	_, _, err := executeCmd([]string{
		"update",
		"--check",
		"--install-go",
		"--github-api", mockURL,
		"--target-path", targetBinary,
	}, "")
	if err == nil {
		t.Fatal("expected error on API failure, got nil")
	}

	// Verify target binary was NOT modified and go install was not invoked
	content, _ := os.ReadFile(targetBinary)
	if !bytes.Equal(content, origContent) {
		t.Errorf("binary was modified during --check --install-go! got: %q, want: %q", content, origContent)
	}
}

func TestCLI_Update_WindowsBackslashChecksumAndReversedFormat(t *testing.T) {
	binaryContent := []byte("backslash-checksum-test\n")
	archiveBytes, err := createTestTarGz("postcode-linux-amd64", binaryContent)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	checksum := computeSHA256(archiveBytes)
	// Checksums file with Windows backslashes and relative paths
	checksums := fmt.Sprintf("%s  dist\\postcode-linux-amd64.tar.gz\n", checksum)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.9.3",
				Assets: []cmd.GitHubAsset{
					{Name: "postcode-linux-amd64.tar.gz", BrowserDownloadURL: mockBase + "/download/archive.tar.gz"},
					{Name: "checksums.txt", BrowserDownloadURL: mockBase + "/download/checksums.txt"},
				},
			}
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/archive.tar.gz":
			_, _ = w.Write(archiveBytes)
		case "/download/checksums.txt":
			_, _ = w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	}))

	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "postcode")
	_ = os.WriteFile(target, []byte("old"), 0755)

	updater := cmd.NewUpdater(cmd.UpdaterConfig{
		CurrentVersion:   "v0.1.0",
		GitHubAPIURL:     mockURL,
		TargetExecutable: target,
		GOOS:             "linux",
		GOARCH:           "amd64",
	})

	res, err := updater.Upgrade(t.Context(), nil, nil)
	if err != nil {
		t.Fatalf("upgrade failed with Windows backslash checksum: %v", err)
	}
	if res.UpdatedVersion != "v0.9.3" {
		t.Errorf("got %s, want v0.9.3", res.UpdatedVersion)
	}
}

func TestCLI_Update_WasmAndTestArtifactRejection(t *testing.T) {
	// Create tar archive with .wasm and test artifacts listed BEFORE the actual executable
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	realBinaryContent := []byte("#!/bin/sh\necho 'real-native-binary'\n")

	files := []struct {
		name    string
		mode    int64
		content []byte
	}{
		{"postcode.wasm", 0644, []byte("\x00asm\x01\x00\x00\x00wasm-binary-data")},
		{"postcode-test", 0755, []byte("test-runner")},
		{"postcode-bench", 0755, []byte("benchmark-suite")},
		{"postcode.pub", 0644, []byte("ssh-ed25519 AAAAC3NzaC1...")},
		{"postcode-darwin-arm64", 0755, realBinaryContent},
	}

	for _, f := range files {
		hdr := &tar.Header{
			Name: f.name,
			Mode: f.mode,
			Size: int64(len(f.content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("write header %s: %v", f.name, err)
		}
		if _, err := tw.Write(f.content); err != nil {
			t.Fatalf("write content %s: %v", f.name, err)
		}
	}
	_ = tw.Close()
	_ = gw.Close()

	archiveBytes := buf.Bytes()
	checksum := computeSHA256(archiveBytes)
	checksums := fmt.Sprintf("%s  postcode-darwin-arm64.tar.gz\n", checksum)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.9.4",
				Assets: []cmd.GitHubAsset{
					{Name: "postcode-darwin-arm64.tar.gz", BrowserDownloadURL: mockBase + "/download/archive.tar.gz"},
					{Name: "checksums.txt", BrowserDownloadURL: mockBase + "/download/checksums.txt"},
				},
			}
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/archive.tar.gz":
			_, _ = w.Write(archiveBytes)
		case "/download/checksums.txt":
			_, _ = w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	}))

	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "postcode")
	_ = os.WriteFile(target, []byte("old"), 0755)

	updater := cmd.NewUpdater(cmd.UpdaterConfig{
		CurrentVersion:   "v0.1.0",
		GitHubAPIURL:     mockURL,
		TargetExecutable: target,
		GOOS:             "darwin",
		GOARCH:           "arm64",
	})

	_, err := updater.Upgrade(t.Context(), nil, nil)
	if err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading updated binary failed: %v", err)
	}
	if !bytes.Equal(data, realBinaryContent) {
		t.Fatalf("wrong artifact extracted instead of real binary! got: %q", string(data))
	}
}

func TestCLI_Update_FallbackAssetDiscovery(t *testing.T) {
	// Release where asset has version string in archive name (postcode-v0.9.5-linux-amd64.tar.gz)
	binaryContent := []byte("versioned-asset-name-binary\n")
	archiveBytes, err := createTestTarGz("postcode-linux-amd64", binaryContent)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	checksum := computeSHA256(archiveBytes)
	versionedArchiveName := "postcode-v0.9.5-linux-amd64.tar.gz"
	checksums := fmt.Sprintf("%s  %s\n", checksum, versionedArchiveName)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.9.5",
				Assets: []cmd.GitHubAsset{
					{Name: versionedArchiveName, BrowserDownloadURL: mockBase + "/download/" + versionedArchiveName},
					{Name: "checksums.txt", BrowserDownloadURL: mockBase + "/download/checksums.txt"},
				},
			}
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/" + versionedArchiveName:
			_, _ = w.Write(archiveBytes)
		case "/download/checksums.txt":
			_, _ = w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	}))

	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "postcode")
	_ = os.WriteFile(target, []byte("old"), 0755)

	updater := cmd.NewUpdater(cmd.UpdaterConfig{
		CurrentVersion:   "v0.1.0",
		GitHubAPIURL:     mockURL,
		TargetExecutable: target,
		GOOS:             "linux",
		GOARCH:           "amd64",
	})

	res, err := updater.Upgrade(t.Context(), nil, nil)
	if err != nil {
		t.Fatalf("upgrade with fallback asset discovery failed: %v", err)
	}
	if res.UpdatedVersion != "v0.9.5" {
		t.Errorf("got %s, want v0.9.5", res.UpdatedVersion)
	}

	data, _ := os.ReadFile(target)
	if !bytes.Equal(data, binaryContent) {
		t.Errorf("content mismatch: got %q, want %q", data, binaryContent)
	}
}

func TestCLI_Update_EmptyCandidateSkip(t *testing.T) {
	// Archive contains an empty candidate entry before the real binary
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	realBinaryContent := []byte("real-payload-after-empty-entry\n")

	files := []struct {
		name    string
		mode    int64
		content []byte
	}{
		{"postcode-darwin-arm64", 0755, []byte{}}, // 0-byte candidate!
		{"postcode", 0755, realBinaryContent},     // Real executable!
	}

	for _, f := range files {
		hdr := &tar.Header{
			Name: f.name,
			Mode: f.mode,
			Size: int64(len(f.content)),
		}
		_ = tw.WriteHeader(hdr)
		_, _ = tw.Write(f.content)
	}
	_ = tw.Close()
	_ = gw.Close()

	archiveBytes := buf.Bytes()
	checksum := computeSHA256(archiveBytes)
	checksums := fmt.Sprintf("%s  postcode-darwin-arm64.tar.gz\n", checksum)

	const mockBase = "http://mock-github-api.local"
	mockURL := registerMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/abcubed3/postcode/releases/latest":
			rel := cmd.GitHubRelease{
				TagName: "v0.9.6",
				Assets: []cmd.GitHubAsset{
					{Name: "postcode-darwin-arm64.tar.gz", BrowserDownloadURL: mockBase + "/download/archive.tar.gz"},
					{Name: "checksums.txt", BrowserDownloadURL: mockBase + "/download/checksums.txt"},
				},
			}
			_ = json.NewEncoder(w).Encode(rel)
		case "/download/archive.tar.gz":
			_, _ = w.Write(archiveBytes)
		case "/download/checksums.txt":
			_, _ = w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	}))

	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "postcode")
	_ = os.WriteFile(target, []byte("old"), 0755)

	updater := cmd.NewUpdater(cmd.UpdaterConfig{
		CurrentVersion:   "v0.1.0",
		GitHubAPIURL:     mockURL,
		TargetExecutable: target,
		GOOS:             "darwin",
		GOARCH:           "arm64",
	})

	res, err := updater.Upgrade(t.Context(), nil, nil)
	if err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if res.UpdatedVersion != "v0.9.6" {
		t.Errorf("got %s, want v0.9.6", res.UpdatedVersion)
	}

	data, _ := os.ReadFile(target)
	if !bytes.Equal(data, realBinaryContent) {
		t.Errorf("target was not updated with real binary payload")
	}
}
