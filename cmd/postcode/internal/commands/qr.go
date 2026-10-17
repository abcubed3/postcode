package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/abcubed3/postcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// QREntry represents a single generated QR code result in CLI structured output.
type QREntry struct {
	Input         string                          `json:"input"`
	Postcode      string                          `json:"postcode"`
	Compact       string                          `json:"compact"`
	StateCode     string                          `json:"state_code"`
	StateName     string                          `json:"state_name"`
	LGACode       string                          `json:"lga_code,omitempty"`
	LGAName       string                          `json:"lga_name,omitempty"`
	Address       string                          `json:"address,omitempty"`
	Latitude      float64                         `json:"latitude"`
	Longitude     float64                         `json:"longitude"`
	GoogleMapsURL string                          `json:"google_maps_url"`
	Payload       string                          `json:"payload"`
	PayloadType   postcode.QRPayloadType          `json:"payload_type"`
	Level         postcode.QRErrorCorrectionLevel `json:"level"`
	Version       int                             `json:"version"`
	Dimension     int                             `json:"dimension"`
	IncludeLabel  bool                            `json:"include_label"`
	RenderFormat  string                          `json:"render_format"`
	SavedFile     string                          `json:"saved_file,omitempty"`
	SVG           string                          `json:"svg,omitempty"`
	DataURI       string                          `json:"data_uri,omitempty"`
	Terminal      string                          `json:"terminal,omitempty"`
}

// QRResult aggregates one or more generated postcode QR codes.
type QRResult struct {
	Total   int       `json:"total"`
	Results []QREntry `json:"results"`
}

func (qr QRResult) CSVHeader() []string {
	return []string{"input", "postcode", "state_name", "lga_name", "payload_type", "payload", "version", "dimension", "saved_file", "data_uri"}
}

func (qr QRResult) CSVRows() [][]string {
	rows := make([][]string, len(qr.Results))
	for i, r := range qr.Results {
		rows[i] = []string{
			r.Input,
			r.Postcode,
			r.StateName,
			r.LGAName,
			string(r.PayloadType),
			r.Payload,
			fmt.Sprintf("%d", r.Version),
			fmt.Sprintf("%d", r.Dimension),
			r.SavedFile,
			r.DataURI,
		}
	}
	return rows
}

// NewQRCmd creates the qr subcommand for converting valid postcodes into shareable or printable QR codes.
func NewQRCmd(v *viper.Viper) *cobra.Command {
	var (
		payload      string
		renderFormat string
		outFile      string
		includeLabel bool
		title        string
		subtitle     string
		size         int
		level        string
		fgColor      string
		bgColor      string
		quietZone    int
		invertASCII  bool
		openViewer   bool
		online       bool
	)

	cmd := &cobra.Command{
		Use:   "qr [postcodes...]",
		Short: "Convert valid postcodes to shareable or printable QR codes (Terminal, SVG, PNG, HTML Label)",
		Long: `QR converts one or more valid 11-character Nigerian postcodes into standards-compliant
(ISO/IEC 18004) QR codes ready for terminal sharing, digital embedding, or high-resolution printing.

By default, the QR code encodes the canonical hyphenated postcode (e.g. EK-01-A03-FK-01).
Use --payload to encode Google Maps navigation URLs, Apple Maps, OpenStreetMap, Geo URIs, or JSON metadata.
Use --label to generate a print-ready NIPOST Postcode Badge with the formatted postcode, State/LGA, and coordinates.`,
		Example: `  # Preview a postcode QR code directly in the terminal
  postcode qr EK-01-A03-FK-01

  # Generate a printable NIPOST postcode badge as an SVG or PNG file
  postcode qr LA-11-W06-TC-10 --label --out label.svg
  postcode qr LA-11-W06-TC-10 --label --out label.png

  # Encode a Google Maps URL inside the QR code and generate a printable HTML sheet
  postcode qr FC-03-B06-AG-12 --payload google-maps --label --out card.html

  # Output structured JSON with SVG and Base64 PNG Data URI for web/API embedding
  postcode qr EK-01-A03-FK-01 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := readInputs(cmd, args)
			if err != nil {
				return err
			}

			useOnline := !isOffline(v, cmd) && online
			var client *postcode.Client
			if useOnline {
				client, _ = buildClient(v)
			}

			effectiveRender := strings.ToLower(strings.TrimSpace(renderFormat))
			renderChanged := cmd.Flags().Changed("render")
			if !renderChanged && outFile != "" {
				ext := strings.ToLower(filepath.Ext(outFile))
				switch ext {
				case ".png":
					effectiveRender = "png"
				case ".svg":
					if includeLabel {
						effectiveRender = "label-svg"
					} else {
						effectiveRender = "svg"
					}
				case ".html", ".htm":
					effectiveRender = "html"
				case ".txt":
					effectiveRender = "terminal"
				}
			}
			if effectiveRender == "" {
				if includeLabel && outFile != "" {
					effectiveRender = "label-svg"
				} else {
					effectiveRender = "terminal"
				}
			}

			switch effectiveRender {
			case "terminal", "ascii", "svg", "label-svg", "html", "png", "data-uri", "svg-data-uri":
			default:
				return fmt.Errorf("invalid render format %q: choose terminal, ascii, svg, label-svg, html, png, data-uri, or svg-data-uri", effectiveRender)
			}

			res := QRResult{
				Total:   len(inputs),
				Results: make([]QREntry, 0, len(inputs)),
			}

			for i, raw := range inputs {
				opts := postcode.QROptions{
					Payload:         postcode.QRPayloadType(payload),
					Level:           postcode.QRErrorCorrectionLevel(level),
					Size:            size,
					QuietZone:       quietZone,
					ForegroundColor: fgColor,
					BackgroundColor: bgColor,
					IncludeLabel:    includeLabel || effectiveRender == "label-svg" || effectiveRender == "html",
					Title:           title,
					Subtitle:        subtitle,
					InvertASCII:     invertASCII,
				}

				var qr *postcode.QRCode
				if useOnline && client != nil {
					qr, err = client.GenerateQR(cmd.Context(), raw, opts)
				} else {
					qr, err = postcode.GenerateQR(raw, opts)
				}
				if err != nil {
					return fmt.Errorf("generating QR for %q: %w", raw, err)
				}

				savedPath := ""
				if outFile != "" {
					targetPath := resolveQROutputPath(outFile, qr.Postcode, effectiveRender, len(inputs) > 1, i)
					if err := writeQRFile(targetPath, qr, effectiveRender); err != nil {
						return fmt.Errorf("saving QR file %s: %w", targetPath, err)
					}
					savedPath = targetPath
					if openViewer {
						_ = openInBrowser(targetPath)
					}
				} else if openViewer {
					// Write temporary printable HTML card and open in browser
					tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("postcode-%s.html", strings.ToLower(qr.Postcode)))
					if err := os.WriteFile(tmpPath, []byte(qr.PrintableHTML()), 0o600); err == nil {
						_ = openInBrowser(tmpPath)
					}
				}

				dataURI, _ := qr.DataURI()
				res.Results = append(res.Results, QREntry{
					Input:         raw,
					Postcode:      qr.Postcode,
					Compact:       qr.Compact,
					StateCode:     qr.StateCode,
					StateName:     qr.StateName,
					LGACode:       qr.LGACode,
					LGAName:       qr.LGAName,
					Address:       qr.Address,
					Latitude:      qr.Latitude,
					Longitude:     qr.Longitude,
					GoogleMapsURL: qr.GoogleMapsURL,
					Payload:       qr.Payload,
					PayloadType:   qr.PayloadType,
					Level:         qr.Level,
					Version:       qr.Version,
					Dimension:     qr.Dimension,
					IncludeLabel:  qr.IncludeLabel,
					RenderFormat:  effectiveRender,
					SavedFile:     savedPath,
					SVG:           qr.SVG(),
					DataURI:       dataURI,
					Terminal:      qr.Terminal(),
				})
			}

			return PrintOutput(cmd, v, res, func(w io.Writer) error {
				for idx, r := range res.Results {
					qr := mustRebuildOrRender(inputs[idx], r, payload, level, size, quietZone, fgColor, bgColor, includeLabel, title, subtitle, invertASCII)
					if r.SavedFile != "" && (effectiveRender == "png" || effectiveRender == "svg" || effectiveRender == "label-svg" || effectiveRender == "html") {
						_, _ = fmt.Fprintf(w, "Saved %s QR (%s, payload=%s) -> %s\n", r.Postcode, effectiveRender, r.PayloadType, r.SavedFile)
						continue
					}
					switch effectiveRender {
					case "svg":
						_, _ = fmt.Fprintln(w, qr.SVG())
					case "label-svg":
						_, _ = fmt.Fprintln(w, qr.LabelSVG())
					case "html":
						_, _ = fmt.Fprintln(w, qr.PrintableHTML())
					case "png":
						pngBytes, err := qr.PNG()
						if err != nil {
							return err
						}
						_, err = w.Write(pngBytes)
						if err != nil {
							return err
						}
					case "data-uri":
						_, _ = fmt.Fprintln(w, r.DataURI)
					case "svg-data-uri":
						_, _ = fmt.Fprintln(w, qr.SVGDataURI())
					case "ascii":
						if !includeLabel {
							_, _ = fmt.Fprintf(w, "%s (%s, %s) -> %s\n", r.Postcode, r.LGAName, r.StateName, r.Payload)
						}
						_, _ = fmt.Fprint(w, qr.ASCII())
					case "terminal":
						fallthrough
					default:
						if !includeLabel {
							_, _ = fmt.Fprintf(w, "%s (%s, %s) -> %s\n", r.Postcode, r.LGAName, r.StateName, r.Payload)
						}
						_, _ = fmt.Fprint(w, qr.Terminal())
					}
				}
				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&payload, "payload", "p", "postcode", "QR payload: postcode, compact, spaced, google-maps, directions, apple-maps, osm, geo, json, or text")
	cmd.Flags().StringVarP(&renderFormat, "render", "r", "terminal", "render format: terminal, ascii, svg, label-svg, html, png, data-uri, or svg-data-uri")
	cmd.Flags().StringVarP(&outFile, "out", "f", "", "write QR output to file (inferring .png, .svg, .html from extension)")
	cmd.Flags().BoolVarP(&includeLabel, "label", "l", false, "render a printable NIPOST postcode badge with postcode, State/LGA, and coordinates")
	cmd.Flags().StringVar(&title, "title", "", "custom header title for printable label (default: 'NIGERIA DIGITAL POSTCODE')")
	cmd.Flags().StringVar(&subtitle, "subtitle", "", "custom street address or subtitle for printable label")
	cmd.Flags().IntVarP(&size, "size", "s", 0, "image size in pixels for PNG/SVG (default 256, or 400 with --label)")
	cmd.Flags().StringVarP(&level, "level", "e", "", "error correction level: L, M, Q, or H (default M, or Q with --label)")
	cmd.Flags().StringVar(&fgColor, "fg", "#000000", "foreground hex color (e.g. #000000 or #008751)")
	cmd.Flags().StringVar(&bgColor, "bg", "#FFFFFF", "background hex color (e.g. #FFFFFF)")
	cmd.Flags().IntVarP(&quietZone, "quiet-zone", "q", 4, "quiet zone margin in modules (set -1 for none)")
	cmd.Flags().BoolVar(&invertASCII, "invert", false, "invert dark/light blocks in terminal/ascii output")
	cmd.Flags().BoolVar(&openViewer, "open", false, "open generated QR file or printable preview in default browser/viewer")
	cmd.Flags().BoolVar(&online, "online", false, "enrich address metadata via NIPOST gateway before generating QR")

	_ = cmd.RegisterFlagCompletionFunc("payload", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"postcode", "compact", "spaced", "google-maps", "directions", "apple-maps", "osm", "geo", "json", "text"}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("render", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"terminal", "ascii", "svg", "label-svg", "html", "png", "data-uri", "svg-data-uri"}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("level", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"L", "M", "Q", "H"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func mustRebuildOrRender(raw string, entry QREntry, payload, level string, size, quietZone int, fg, bg string, includeLabel bool, title, subtitle string, invert bool) *postcode.QRCode {
	loc := postcode.Location{
		Postcode:  entry.Postcode,
		Compact:   entry.Compact,
		Latitude:  entry.Latitude,
		Longitude: entry.Longitude,
		Address:   entry.Address,
		StateCode: entry.StateCode,
		StateName: entry.StateName,
		LGACode:   entry.LGACode,
		LGAName:   entry.LGAName,
	}
	qr, err := loc.QR(postcode.QROptions{
		Payload:         postcode.QRPayloadType(payload),
		Level:           postcode.QRErrorCorrectionLevel(level),
		Size:            size,
		QuietZone:       quietZone,
		ForegroundColor: fg,
		BackgroundColor: bg,
		IncludeLabel:    includeLabel,
		Title:           title,
		Subtitle:        subtitle,
		InvertASCII:     invert,
	})
	if err != nil {
		qr, _ = postcode.GenerateQR(raw)
	}
	return qr
}

func resolveQROutputPath(outFile, formattedPostcode, renderFormat string, multiple bool, index int) string {
	slug := strings.ToLower(formattedPostcode)
	if strings.Contains(outFile, "{postcode}") {
		return strings.ReplaceAll(outFile, "{postcode}", formattedPostcode)
	}
	if strings.Contains(outFile, "{slug}") {
		return strings.ReplaceAll(outFile, "{slug}", slug)
	}

	defaultExt := ".txt"
	switch renderFormat {
	case "png":
		defaultExt = ".png"
	case "svg", "label-svg":
		defaultExt = ".svg"
	case "html":
		defaultExt = ".html"
	}

	if info, err := os.Stat(outFile); err == nil && info.IsDir() {
		return filepath.Join(outFile, slug+defaultExt)
	}
	if strings.HasSuffix(outFile, "/") || strings.HasSuffix(outFile, string(os.PathSeparator)) {
		return filepath.Join(outFile, slug+defaultExt)
	}

	if !multiple {
		return outFile
	}

	ext := filepath.Ext(outFile)
	base := strings.TrimSuffix(outFile, ext)
	if ext == "" {
		ext = defaultExt
	}
	return fmt.Sprintf("%s-%s%s", base, slug, ext)
}

func writeQRFile(path string, qr *postcode.QRCode, renderFormat string) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	switch renderFormat {
	case "png":
		pngBytes, err := qr.PNG()
		if err != nil {
			return err
		}
		return os.WriteFile(path, pngBytes, 0o644)
	case "label-svg":
		return os.WriteFile(path, []byte(qr.LabelSVG()), 0o644)
	case "svg":
		return os.WriteFile(path, []byte(qr.SVG()), 0o644)
	case "html":
		return os.WriteFile(path, []byte(qr.PrintableHTML()), 0o644)
	case "data-uri":
		uri, err := qr.DataURI()
		if err != nil {
			return err
		}
		return os.WriteFile(path, []byte(uri+"\n"), 0o644)
	case "svg-data-uri":
		return os.WriteFile(path, []byte(qr.SVGDataURI()+"\n"), 0o644)
	case "ascii":
		return os.WriteFile(path, []byte(qr.ASCII()), 0o644)
	case "terminal":
		fallthrough
	default:
		return os.WriteFile(path, []byte(qr.Terminal()), 0o644)
	}
}
