package postcode

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// QRErrorCorrectionLevel defines the Reed-Solomon error correction capacity of a QR code.
type QRErrorCorrectionLevel string

const (
	// QRLevelL recovers up to ~7% of damaged codewords.
	QRLevelL QRErrorCorrectionLevel = "L"
	// QRLevelM recovers up to ~15% of damaged codewords (default).
	QRLevelM QRErrorCorrectionLevel = "M"
	// QRLevelQ recovers up to ~25% of damaged codewords (recommended for printed labels).
	QRLevelQ QRErrorCorrectionLevel = "Q"
	// QRLevelH recovers up to ~30% of damaged codewords.
	QRLevelH QRErrorCorrectionLevel = "H"
)

// QRPayloadType specifies what data is encoded inside the QR code when scanned.
type QRPayloadType string

const (
	// QRPayloadPostcode encodes the canonical hyphenated 11-character postcode (e.g. "EK-01-A03-FK-01"). Default.
	QRPayloadPostcode QRPayloadType = "postcode"
	// QRPayloadCompact encodes the compact 11-character postcode without hyphens (e.g. "EK01A03FK01").
	QRPayloadCompact QRPayloadType = "compact"
	// QRPayloadSpaced encodes the spaced 5-segment display format (e.g. "EK 01 A03 FK 01").
	QRPayloadSpaced QRPayloadType = "spaced"
	// QRPayloadGoogleMaps encodes a direct Google Maps search URL for the postcode location.
	QRPayloadGoogleMaps QRPayloadType = "google-maps"
	// QRPayloadDirections encodes a turn-by-turn Google Maps navigation URL to the postcode.
	QRPayloadDirections QRPayloadType = "directions"
	// QRPayloadAppleMaps encodes an Apple Maps pin URL for iOS/macOS devices.
	QRPayloadAppleMaps QRPayloadType = "apple-maps"
	// QRPayloadOSM encodes an OpenStreetMap pin URL.
	QRPayloadOSM QRPayloadType = "osm"
	// QRPayloadGeoURI encodes an RFC 5870 geo: URI (e.g. "geo:7.621100,5.221500?q=EK-01-A03-FK-01").
	QRPayloadGeoURI QRPayloadType = "geo"
	// QRPayloadJSON encodes a compact JSON object with postcode, state, LGA, coordinates, and map URL.
	QRPayloadJSON QRPayloadType = "json"
	// QRPayloadText encodes a human-readable multi-line address card with postcode, location, and map link.
	QRPayloadText QRPayloadType = "text"
)

// QROptions configures how a postcode QR code is encoded and rendered for sharing or printing.
type QROptions struct {
	// Payload controls what content the QR code encodes when scanned. Defaults to QRPayloadPostcode ("postcode").
	Payload QRPayloadType `json:"payload,omitempty"`
	// Level sets the Reed-Solomon error correction level ("L", "M", "Q", "H"). Defaults to "M" (or "Q" when IncludeLabel is true).
	Level QRErrorCorrectionLevel `json:"level,omitempty"`
	// Size is the target width/height in pixels for SVG and PNG rendering. Defaults to 256 (or 400 when IncludeLabel is true).
	Size int `json:"size,omitempty"`
	// QuietZone is the margin around the QR matrix in modules. Defaults to 4. Set to -1 for zero margin.
	QuietZone int `json:"quiet_zone,omitempty"`
	// ForegroundColor is the hex color for dark QR modules (e.g. "#000000" or "#008751"). Defaults to "#000000".
	ForegroundColor string `json:"foreground_color,omitempty"`
	// BackgroundColor is the hex color for light QR background (e.g. "#FFFFFF"). Defaults to "#FFFFFF".
	BackgroundColor string `json:"background_color,omitempty"`
	// IncludeLabel renders a printable NIPOST postcode badge/label with human-readable postcode, State/LGA, and coordinates.
	IncludeLabel bool `json:"include_label,omitempty"`
	// Title overrides the header text on printable labels (defaults to "NIGERIA DIGITAL POSTCODE").
	Title string `json:"title,omitempty"`
	// Subtitle adds a custom street address or recipient line on printable labels.
	Subtitle string `json:"subtitle,omitempty"`
	// InvertASCII inverts dark and light blocks when rendering terminal/ASCII output.
	InvertASCII bool `json:"invert_ascii,omitempty"`
}

// QRCode represents a generated, standards-compliant (ISO/IEC 18004) QR code for a Nigerian postcode.
type QRCode struct {
	Postcode        string                 `json:"postcode"`
	Compact         string                 `json:"compact"`
	Spaced          string                 `json:"spaced"`
	StateCode       string                 `json:"state_code"`
	StateName       string                 `json:"state_name"`
	LGACode         string                 `json:"lga_code,omitempty"`
	LGAName         string                 `json:"lga_name,omitempty"`
	Zone            string                 `json:"zone,omitempty"`
	Address         string                 `json:"address,omitempty"`
	Latitude        float64                `json:"latitude"`
	Longitude       float64                `json:"longitude"`
	GoogleMapsURL   string                 `json:"google_maps_url"`
	Payload         string                 `json:"payload"`
	PayloadType     QRPayloadType          `json:"payload_type"`
	Level           QRErrorCorrectionLevel `json:"level"`
	Version         int                    `json:"version"`
	Dimension       int                    `json:"dimension"`
	QuietZone       int                    `json:"quiet_zone"`
	Size            int                    `json:"size"`
	ForegroundColor string                 `json:"foreground_color"`
	BackgroundColor string                 `json:"background_color"`
	IncludeLabel    bool                   `json:"include_label"`
	Title           string                 `json:"title,omitempty"`
	Subtitle        string                 `json:"subtitle,omitempty"`
	InvertASCII     bool                   `json:"invert_ascii,omitempty"`
	Modules         [][]bool               `json:"-"`
}

// GeoURI returns an RFC 5870 geo: URI for this location (e.g. "geo:7.621100,5.221500?q=EK-01-A03-FK-01").
func (l Location) GeoURI() string {
	if l.Latitude == 0 && l.Longitude == 0 {
		if l.Postcode != "" {
			return fmt.Sprintf("geo:0,0?q=%s", l.Postcode)
		}
		return ""
	}
	if l.Postcode != "" {
		return fmt.Sprintf("geo:%.6f,%.6f?q=%s", l.Latitude, l.Longitude, l.Postcode)
	}
	return fmt.Sprintf("geo:%.6f,%.6f", l.Latitude, l.Longitude)
}

// GenerateQR validates a Nigerian postcode string and generates a QRCode with the given options.
func GenerateQR(raw string, opts ...QROptions) (*QRCode, error) {
	if err := Validate(raw); err != nil {
		return nil, err
	}
	p, err := Parse(raw)
	if err != nil {
		return nil, err
	}
	return p.Location().QR(opts...)
}

// QRSVG validates a Nigerian postcode string and returns its SVG QR code representation.
func QRSVG(raw string, opts ...QROptions) (string, error) {
	qr, err := GenerateQR(raw, opts...)
	if err != nil {
		return "", err
	}
	return qr.SVG(), nil
}

// QRPNG validates a Nigerian postcode string and returns its PNG QR code image bytes.
func QRPNG(raw string, opts ...QROptions) ([]byte, error) {
	qr, err := GenerateQR(raw, opts...)
	if err != nil {
		return nil, err
	}
	return qr.PNG()
}

// QRTerminal validates a Nigerian postcode string and returns a compact Unicode terminal QR code.
func QRTerminal(raw string, opts ...QROptions) (string, error) {
	qr, err := GenerateQR(raw, opts...)
	if err != nil {
		return "", err
	}
	return qr.Terminal(), nil
}

// QR generates a QRCode for this Postcode.
// Returns an error if the postcode is zero or has an unrecognized Nigerian state code.
func (p Postcode) QR(opts ...QROptions) (*QRCode, error) {
	if p.IsZero() {
		return nil, ErrInvalidLength
	}
	if _, ok := NigerianStates[p.State()]; !ok {
		return nil, fmt.Errorf("%w: unknown state code %q (not in 36 States + FCT)", ErrInvalidFormat, p.State())
	}
	return p.Location().QR(opts...)
}

// QRSVG returns an SVG string of the QR code for this Postcode.
func (p Postcode) QRSVG(opts ...QROptions) (string, error) {
	qr, err := p.QR(opts...)
	if err != nil {
		return "", err
	}
	return qr.SVG(), nil
}

// QRPNG returns PNG bytes of the QR code for this Postcode.
func (p Postcode) QRPNG(opts ...QROptions) ([]byte, error) {
	qr, err := p.QR(opts...)
	if err != nil {
		return nil, err
	}
	return qr.PNG()
}

// QRTerminal returns a compact Unicode terminal representation of the QR code for this Postcode.
func (p Postcode) QRTerminal(opts ...QROptions) (string, error) {
	qr, err := p.QR(opts...)
	if err != nil {
		return "", err
	}
	return qr.Terminal(), nil
}

// GenerateQR resolves the postcode location via the NIPOST gateway (falling back to offline reference data)
// and generates a QRCode enriched with any available street address and point coordinates.
func (c *Client) GenerateQR(ctx context.Context, code string, opts ...QROptions) (*QRCode, error) {
	if err := Validate(code); err != nil {
		return nil, err
	}
	loc, err := c.ResolveLocation(ctx, code)
	if err != nil {
		offlineLoc, offErr := ResolveLocation(code)
		if offErr != nil {
			return nil, offErr
		}
		loc = &offlineLoc
	}
	return loc.QR(opts...)
}

// QR generates a QRCode from a resolved Location, preserving any enriched address and coordinate data.
func (l Location) QR(opts ...QROptions) (*QRCode, error) {
	if l.Postcode == "" {
		return nil, ErrInvalidLength
	}
	p, err := Parse(l.Postcode)
	if err != nil {
		return nil, err
	}
	if _, ok := NigerianStates[p.State()]; !ok {
		return nil, fmt.Errorf("%w: unknown state code %q (not in 36 States + FCT)", ErrInvalidFormat, p.State())
	}

	var opt QROptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	payloadType, err := normalizeQRPayloadType(opt.Payload)
	if err != nil {
		return nil, err
	}

	level, err := normalizeQRLevel(opt.Level, opt.IncludeLabel)
	if err != nil {
		return nil, err
	}

	size := opt.Size
	if size <= 0 {
		if opt.IncludeLabel {
			size = 400
		} else {
			size = 256
		}
	}

	quietZone := opt.QuietZone
	if quietZone == 0 {
		quietZone = 4
	} else if quietZone < 0 {
		quietZone = 0
	}

	fg := normalizeHexColor(opt.ForegroundColor, "#000000")
	bg := normalizeHexColor(opt.BackgroundColor, "#FFFFFF")

	title := strings.TrimSpace(opt.Title)
	if title == "" {
		title = "NIGERIA DIGITAL POSTCODE"
	}
	subtitle := strings.TrimSpace(opt.Subtitle)
	if subtitle == "" && l.Address != "" {
		subtitle = l.Address
	}

	payloadStr := buildQRPayload(p, l, payloadType)
	modules, ver, err := encodeQRMatrix([]byte(payloadStr), level)
	if err != nil {
		return nil, err
	}

	return &QRCode{
		Postcode:        p.Formatted(),
		Compact:         p.Compact(),
		Spaced:          p.String(),
		StateCode:       p.State(),
		StateName:       l.StateName,
		LGACode:         p.LGA(),
		LGAName:         l.LGAName,
		Zone:            l.Zone,
		Address:         l.Address,
		Latitude:        l.Latitude,
		Longitude:       l.Longitude,
		GoogleMapsURL:   l.GoogleMapsURL(),
		Payload:         payloadStr,
		PayloadType:     payloadType,
		Level:           level,
		Version:         ver,
		Dimension:       len(modules),
		QuietZone:       quietZone,
		Size:            size,
		ForegroundColor: fg,
		BackgroundColor: bg,
		IncludeLabel:    opt.IncludeLabel,
		Title:           title,
		Subtitle:        subtitle,
		InvertASCII:     opt.InvertASCII,
		Modules:         modules,
	}, nil
}

func normalizeQRPayloadType(pt QRPayloadType) (QRPayloadType, error) {
	s := strings.ToLower(strings.TrimSpace(string(pt)))
	switch s {
	case "", "postcode", "canonical", "hyphenated", "code":
		return QRPayloadPostcode, nil
	case "compact", "raw":
		return QRPayloadCompact, nil
	case "spaced", "display":
		return QRPayloadSpaced, nil
	case "google-maps", "google_maps", "google", "gmaps", "map", "url":
		return QRPayloadGoogleMaps, nil
	case "directions", "navigation", "nav", "route":
		return QRPayloadDirections, nil
	case "apple-maps", "apple_maps", "apple":
		return QRPayloadAppleMaps, nil
	case "osm", "openstreetmap":
		return QRPayloadOSM, nil
	case "geo", "geouri", "geo-uri", "coords", "coordinates":
		return QRPayloadGeoURI, nil
	case "json", "metadata":
		return QRPayloadJSON, nil
	case "text", "share", "card":
		return QRPayloadText, nil
	default:
		return "", fmt.Errorf("postcode: unsupported QR payload type %q (choose postcode, compact, spaced, google-maps, directions, apple-maps, osm, geo, json, or text)", pt)
	}
}

func normalizeQRLevel(lvl QRErrorCorrectionLevel, includeLabel bool) (QRErrorCorrectionLevel, error) {
	s := strings.ToUpper(strings.TrimSpace(string(lvl)))
	switch s {
	case "":
		if includeLabel {
			return QRLevelQ, nil
		}
		return QRLevelM, nil
	case "L", "LOW":
		return QRLevelL, nil
	case "M", "MEDIUM":
		return QRLevelM, nil
	case "Q", "QUARTILE":
		return QRLevelQ, nil
	case "H", "HIGH":
		return QRLevelH, nil
	default:
		return "", fmt.Errorf("postcode: invalid QR error correction level %q (choose L, M, Q, or H)", lvl)
	}
}

func normalizeHexColor(c, fallback string) string {
	c = strings.TrimSpace(c)
	if c == "" {
		return fallback
	}
	if !strings.HasPrefix(c, "#") {
		c = "#" + c
	}
	if len(c) == 4 {
		// Expand #RGB to #RRGGBB
		r, g, b := c[1], c[2], c[3]
		c = string([]byte{'#', r, r, g, g, b, b})
	}
	if len(c) != 7 {
		return fallback
	}
	for i := 1; i < 7; i++ {
		ch := c[i]
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
			return fallback
		}
	}
	return strings.ToUpper(c)
}

func parseHexRGBA(hexStr string, fallback color.RGBA) color.RGBA {
	if len(hexStr) != 7 || hexStr[0] != '#' {
		return fallback
	}
	rv, err1 := strconv.ParseUint(hexStr[1:3], 16, 8)
	gv, err2 := strconv.ParseUint(hexStr[3:5], 16, 8)
	bv, err3 := strconv.ParseUint(hexStr[5:7], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return fallback
	}
	return color.RGBA{R: uint8(rv), G: uint8(gv), B: uint8(bv), A: 255}
}

func buildQRPayload(p Postcode, loc Location, pt QRPayloadType) string {
	switch pt {
	case QRPayloadCompact:
		return p.Compact()
	case QRPayloadSpaced:
		return p.String()
	case QRPayloadGoogleMaps:
		return loc.GoogleMapsURL()
	case QRPayloadDirections:
		return loc.GoogleMapsDirectionsURL()
	case QRPayloadAppleMaps:
		return loc.AppleMapsURL()
	case QRPayloadOSM:
		return loc.OpenStreetMapURL()
	case QRPayloadGeoURI:
		return loc.GeoURI()
	case QRPayloadJSON:
		payload := map[string]any{
			"postcode":        p.Formatted(),
			"compact":         p.Compact(),
			"state_code":      p.State(),
			"state_name":      loc.StateName,
			"lga_code":        p.LGA(),
			"lga_name":        loc.LGAName,
			"latitude":        loc.Latitude,
			"longitude":       loc.Longitude,
			"google_maps_url": loc.GoogleMapsURL(),
		}
		if loc.Address != "" {
			payload["address"] = loc.Address
		}
		b, _ := json.Marshal(payload)
		return string(b)
	case QRPayloadText:
		var lines []string
		lines = append(lines, fmt.Sprintf("NIPOST Postcode: %s", p.Formatted()))
		if loc.Address != "" {
			lines = append(lines, fmt.Sprintf("Address: %s", loc.Address))
		}
		if loc.LGAName != "" && loc.StateName != "" {
			lines = append(lines, fmt.Sprintf("Area: %s, %s State", loc.LGAName, loc.StateName))
		} else if loc.StateName != "" {
			lines = append(lines, fmt.Sprintf("State: %s", loc.StateName))
		}
		if loc.Latitude != 0 || loc.Longitude != 0 {
			lines = append(lines, fmt.Sprintf("Coords: %.6f, %.6f", loc.Latitude, loc.Longitude))
		}
		if gurl := loc.GoogleMapsURL(); gurl != "" {
			lines = append(lines, fmt.Sprintf("Map: %s", gurl))
		}
		return strings.Join(lines, "\n")
	case QRPayloadPostcode:
		fallthrough
	default:
		return p.Formatted()
	}
}

// SVG returns a scalable vector graphics (<svg>) representation of the QR code.
// If IncludeLabel was set in QROptions, it returns the full printable NIPOST postcode badge SVG.
func (q *QRCode) SVG() string {
	if q.IncludeLabel {
		return q.LabelSVG()
	}
	return q.MatrixSVG()
}

// MatrixSVG returns a standalone SVG of the QR code matrix without an outer printable card frame.
func (q *QRCode) MatrixSVG() string {
	n := len(q.Modules)
	qz := q.QuietZone
	total := n + 2*qz
	size := q.Size
	if size <= 0 {
		size = 256
	}

	var sb strings.Builder
	sb.Grow(4096)
	fmt.Fprintf(&sb, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" shape-rendering="crispEdges" role="img" aria-label="Nigerian Postcode %s QR Code">`,
		total, total, size, size, html.EscapeString(q.Postcode))
	fmt.Fprintf(&sb, `<title>Nigerian Postcode %s</title>`, html.EscapeString(q.Postcode))
	fmt.Fprintf(&sb, `<rect width="%d" height="%d" fill="%s"/>`, total, total, q.BackgroundColor)
	fmt.Fprintf(&sb, `<path fill="%s" d="`, q.ForegroundColor)
	writeSVGModulesPath(&sb, q.Modules, qz, qz)
	sb.WriteString(`"/></svg>`)
	return sb.String()
}

func writeSVGModulesPath(sb *strings.Builder, modules [][]bool, offsetX, offsetY int) {
	for r, row := range modules {
		y := r + offsetY
		c := 0
		for c < len(row) {
			if !row[c] {
				c++
				continue
			}
			// Merge horizontal contiguous dark modules into a single rectangle M x y hW v1 h-W z
			startC := c
			for c < len(row) && row[c] {
				c++
			}
			w := c - startC
			x := startC + offsetX
			fmt.Fprintf(sb, "M%d %dh%dv1h-%dz", x, y, w, w)
		}
	}
}

// LabelSVG renders a print-ready NIPOST Postcode Badge / Placard SVG containing the QR code,
// canonical hyphenated postcode, State & LGA administrative details, and coordinates.
func (q *QRCode) LabelSVG() string {
	cardW := 400
	cardH := 530
	if q.Subtitle != "" {
		cardH = 555
	}

	headerColor := "#008751" // Official Nigerian / NIPOST Green
	if q.ForegroundColor != "#000000" && q.ForegroundColor != "" {
		headerColor = q.ForegroundColor
	}

	n := len(q.Modules)
	qz := 2
	qrTotal := n + 2*qz
	qrBoxSize := 250
	qrX := (cardW - qrBoxSize) / 2
	qrY := 68

	adminLine := q.StateName
	if q.LGAName != "" && q.StateName != "" {
		if q.StateCode == "FC" {
			adminLine = fmt.Sprintf("%s, %s", q.LGAName, q.StateName)
		} else {
			adminLine = fmt.Sprintf("%s, %s State", q.LGAName, q.StateName)
		}
	}

	scanHint := "SCAN FOR DIGITAL POSTCODE"
	switch q.PayloadType {
	case QRPayloadGoogleMaps, QRPayloadDirections, QRPayloadAppleMaps, QRPayloadOSM, QRPayloadGeoURI:
		scanHint = "SCAN TO OPEN MAP NAVIGATION"
	case QRPayloadJSON:
		scanHint = "SCAN FOR POSTCODE METADATA"
	case QRPayloadText:
		scanHint = "SCAN FOR ADDRESS DETAILS"
	}

	title := q.Title
	if title == "" {
		title = "NIGERIA DIGITAL POSTCODE"
	}

	var sb strings.Builder
	sb.Grow(8192)
	fmt.Fprintf(&sb, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" role="img" aria-label="Printable Nigerian Postcode Label %s">`,
		cardW, cardH, cardW, cardH, html.EscapeString(q.Postcode))
	fmt.Fprintf(&sb, `<title>%s - %s</title>`, html.EscapeString(title), html.EscapeString(q.Postcode))

	// Card background & outer border
	fmt.Fprintf(&sb, `<rect x="2" y="2" width="%d" height="%d" rx="16" ry="16" fill="%s" stroke="%s" stroke-width="3"/>`,
		cardW-4, cardH-4, q.BackgroundColor, headerColor)

	// Top header banner
	fmt.Fprintf(&sb, `<path d="M 2 18 Q 2 2 18 2 L %d 2 Q %d 2 %d 18 L %d 50 L 2 50 Z" fill="%s"/>`,
		cardW-18, cardW-2, cardW-2, cardW-2, headerColor)
	fmt.Fprintf(&sb, `<text x="%d" y="32" text-anchor="middle" fill="#FFFFFF" font-family="system-ui, -apple-system, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif" font-size="15" font-weight="800" letter-spacing="1.8">%s</text>`,
		cardW/2, html.EscapeString(strings.ToUpper(title)))

	// QR code nested viewport
	fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="%d" rx="8" ry="8" fill="%s" stroke="#E2E8F0" stroke-width="1.5"/>`,
		qrX-6, qrY-6, qrBoxSize+12, qrBoxSize+12, q.BackgroundColor)
	fmt.Fprintf(&sb, `<svg x="%d" y="%d" width="%d" height="%d" viewBox="0 0 %d %d" shape-rendering="crispEdges">`,
		qrX, qrY, qrBoxSize, qrBoxSize, qrTotal, qrTotal)
	fmt.Fprintf(&sb, `<rect width="%d" height="%d" fill="%s"/>`, qrTotal, qrTotal, q.BackgroundColor)
	fmt.Fprintf(&sb, `<path fill="%s" d="`, q.ForegroundColor)
	writeSVGModulesPath(&sb, q.Modules, qz, qz)
	sb.WriteString(`"/></svg>`)

	// Canonical Postcode Pill
	pillY := qrY + qrBoxSize + 18
	fmt.Fprintf(&sb, `<rect x="32" y="%d" width="%d" height="54" rx="10" ry="10" fill="#F0FDF4" stroke="%s" stroke-width="2"/>`,
		pillY, cardW-64, headerColor)
	fmt.Fprintf(&sb, `<text x="%d" y="%d" text-anchor="middle" fill="#0F172A" font-family="ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace" font-size="28" font-weight="900" letter-spacing="2">%s</text>`,
		cardW/2, pillY+36, html.EscapeString(q.Postcode))

	// Segment Legend
	legendY := pillY + 72
	fmt.Fprintf(&sb, `<text x="%d" y="%d" text-anchor="middle" fill="#64748B" font-family="system-ui, -apple-system, sans-serif" font-size="10" font-weight="700" letter-spacing="1.5">STATE (%s) • LGA (%s) • DIST (%s) • AREA (%s) • UNIT (%s)</text>`,
		cardW/2, legendY,
		html.EscapeString(q.StateCode),
		html.EscapeString(q.LGACode),
		html.EscapeString(q.Postcode[6:9]),
		html.EscapeString(q.Postcode[10:12]),
		html.EscapeString(q.Postcode[13:15]),
	)

	// Administrative Location Line
	locY := legendY + 26
	if adminLine != "" {
		fmt.Fprintf(&sb, `<text x="%d" y="%d" text-anchor="middle" fill="#1E293B" font-family="system-ui, -apple-system, sans-serif" font-size="15" font-weight="700">%s</text>`,
			cardW/2, locY, html.EscapeString(adminLine))
	}

	nextY := locY
	if q.Subtitle != "" {
		nextY += 22
		sub := q.Subtitle
		if len(sub) > 48 {
			sub = sub[:45] + "..."
		}
		fmt.Fprintf(&sb, `<text x="%d" y="%d" text-anchor="middle" fill="#475569" font-family="system-ui, -apple-system, sans-serif" font-size="12" font-weight="500">%s</text>`,
			cardW/2, nextY, html.EscapeString(sub))
	}

	// Coordinates & Scan Hint Footer
	if q.Latitude != 0 || q.Longitude != 0 {
		nextY += 20
		fmt.Fprintf(&sb, `<text x="%d" y="%d" text-anchor="middle" fill="#64748B" font-family="ui-monospace, SFMono-Regular, Menlo, monospace" font-size="11">%.6f° N, %.6f° E</text>`,
			cardW/2, nextY, q.Latitude, q.Longitude)
	}

	fmt.Fprintf(&sb, `<text x="%d" y="%d" text-anchor="middle" fill="%s" font-family="system-ui, -apple-system, sans-serif" font-size="10" font-weight="800" letter-spacing="1.4">%s</text>`,
		cardW/2, cardH-16, headerColor, html.EscapeString(scanHint))

	sb.WriteString(`</svg>`)
	return sb.String()
}

// PrintableHTML generates a standalone, print-ready HTML page containing the NIPOST Postcode QR Card
// with one-click printing and sharing metadata.
func (q *QRCode) PrintableHTML() string {
	svgCard := q.LabelSVG()
	admin := q.StateName
	if q.LGAName != "" && q.StateName != "" {
		admin = fmt.Sprintf("%s, %s", q.LGAName, q.StateName)
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Nigerian Postcode %s - QR Label</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0;
    padding: 32px 16px;
    background: #F8FAFC;
    color: #0F172A;
    font-family: system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif;
    display: flex;
    flex-direction: column;
    align-items: center;
  }
  .toolbar {
    margin-bottom: 24px;
    display: flex;
    gap: 12px;
  }
  .btn {
    background: #008751;
    color: #FFFFFF;
    border: none;
    border-radius: 8px;
    padding: 10px 20px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    text-decoration: none;
  }
  .btn:hover { background: #006B3F; }
  .card-container {
    background: #FFFFFF;
    padding: 16px;
    border-radius: 20px;
    box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.08);
  }
  .meta {
    margin-top: 20px;
    font-size: 13px;
    color: #475569;
    text-align: center;
    max-width: 400px;
    word-break: break-all;
  }
  @media print {
    body { background: #FFFFFF; padding: 0; }
    .toolbar, .meta { display: none; }
    .card-container { box-shadow: none; padding: 0; }
  }
</style>
</head>
<body>
  <div class="toolbar">
    <button class="btn" onclick="window.print()">Print Postcode Label</button>
  </div>
  <div class="card-container">
    %s
  </div>
  <div class="meta">
    <strong>Postcode:</strong> %s (%s)<br>
    <strong>Encoded Payload (%s):</strong> %s
  </div>
</body>
</html>`,
		html.EscapeString(q.Postcode),
		svgCard,
		html.EscapeString(q.Postcode),
		html.EscapeString(admin),
		html.EscapeString(string(q.PayloadType)),
		html.EscapeString(q.Payload),
	)
}

// Image renders the QRCode as a standard Go image.Image.
// When IncludeLabel is true, a footer banner with the formatted postcode and State/LGA is drawn below the QR matrix.
func (q *QRCode) Image() image.Image {
	n := len(q.Modules)
	qz := q.QuietZone
	totalModules := n + 2*qz
	if totalModules <= 0 {
		totalModules = 1
	}

	targetSize := q.Size
	if targetSize <= 0 {
		targetSize = 256
	}

	scale := targetSize / totalModules
	if scale < 4 {
		scale = 4
	}
	qrPixSize := totalModules * scale

	labelHeight := 0
	if q.IncludeLabel {
		labelHeight = max(56, qrPixSize/5)
	}

	fgColor := parseHexRGBA(q.ForegroundColor, color.RGBA{0, 0, 0, 255})
	bgColor := parseHexRGBA(q.BackgroundColor, color.RGBA{255, 255, 255, 255})
	accentColor := color.RGBA{0, 135, 81, 255} // NIPOST Green
	if q.ForegroundColor != "#000000" {
		accentColor = fgColor
	}

	palette := color.Palette{bgColor, fgColor, accentColor}
	img := image.NewPaletted(image.Rect(0, 0, qrPixSize, qrPixSize+labelHeight), palette)

	// Fill background (palette index 0)
	for i := range img.Pix {
		img.Pix[i] = 0
	}

	// Draw QR modules (palette index 1)
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			if !q.Modules[r][c] {
				continue
			}
			startX := (c + qz) * scale
			startY := (r + qz) * scale
			for dy := 0; dy < scale; dy++ {
				rowOff := (startY+dy)*img.Stride + startX
				for dx := 0; dx < scale; dx++ {
					img.Pix[rowOff+dx] = 1
				}
			}
		}
	}

	if q.IncludeLabel {
		// Draw divider bar in accent color (palette index 2)
		barY := qrPixSize
		for dy := 0; dy < max(2, scale/2); dy++ {
			if barY+dy < qrPixSize+labelHeight {
				rowOff := (barY + dy) * img.Stride
				for x := scale; x < qrPixSize-scale; x++ {
					img.Pix[rowOff+x] = 2
				}
			}
		}

		// Draw canonical postcode centered using built-in 5x7 bitmap font
		fontScale := max(2, qrPixSize/110)
		textY := qrPixSize + max(8, labelHeight/6)
		drawBitmapTextCentered(img, q.Postcode, qrPixSize/2, textY, fontScale, 1)

		// Draw State / LGA line below postcode
		subText := strings.ToUpper(q.StateName)
		if q.LGAName != "" && q.StateName != "" {
			subText = strings.ToUpper(q.LGAName + " - " + q.StateName)
		}
		if len(subText) > 26 {
			subText = subText[:26]
		}
		subScale := max(1, fontScale-1)
		subY := textY + 7*fontScale + max(6, fontScale*2)
		if subY+7*subScale < qrPixSize+labelHeight-2 {
			drawBitmapTextCentered(img, subText, qrPixSize/2, subY, subScale, 2)
		}
	}

	return img
}

// PNG encodes the QR code as a PNG image.
func (q *QRCode) PNG() ([]byte, error) {
	var buf bytes.Buffer
	if err := q.WritePNG(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WritePNG writes the PNG-encoded QR code to w.
func (q *QRCode) WritePNG(w io.Writer) error {
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	return enc.Encode(w, q.Image())
}

// WriteSVG writes the SVG-encoded QR code to w.
func (q *QRCode) WriteSVG(w io.Writer) error {
	_, err := io.WriteString(w, q.SVG())
	return err
}

// SaveSVG writes the SVG-encoded QR code to a file at filename.
func (q *QRCode) SaveSVG(filename string) error {
	return os.WriteFile(filename, []byte(q.SVG()), 0644)
}

// SavePNG writes the PNG-encoded QR code image to a file at filename.
func (q *QRCode) SavePNG(filename string) error {
	b, err := q.PNG()
	if err != nil {
		return err
	}
	return os.WriteFile(filename, b, 0644)
}

// SaveHTML writes the printable HTML label page to a file at filename.
func (q *QRCode) SaveHTML(filename string) error {
	return os.WriteFile(filename, []byte(q.PrintableHTML()), 0644)
}

// DataURI returns a base64-encoded PNG data URI ("data:image/png;base64,...") suitable for HTML <img> src attributes.
func (q *QRCode) DataURI() (string, error) {
	pngBytes, err := q.PNG()
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes), nil
}

// SVGDataURI returns a base64-encoded SVG data URI ("data:image/svg+xml;base64,...").
func (q *QRCode) SVGDataURI() string {
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(q.SVG()))
}

// Terminal renders the QR code using compact Unicode half-block characters ('▀', '▄', '█', ' ')
// so that two vertical modules occupy a single terminal row, preserving a 1:1 square aspect ratio.
func (q *QRCode) Terminal() string {
	n := len(q.Modules)
	qz := q.QuietZone
	if qz > 2 {
		qz = 2 // Keep terminal output compact while preserving scan quiet zone
	}
	total := n + 2*qz

	isDark := func(r, c int) bool {
		mr := r - qz
		mc := c - qz
		var dark bool
		if mr >= 0 && mr < n && mc >= 0 && mc < n {
			dark = q.Modules[mr][mc]
		}
		if q.InvertASCII {
			return !dark
		}
		return dark
	}

	var sb strings.Builder
	if q.IncludeLabel {
		fmt.Fprintf(&sb, "┌─ %s (%s) ─┐\n", q.Postcode, q.StateName)
	}

	for r := 0; r < total; r += 2 {
		for c := 0; c < total; c++ {
			top := isDark(r, c)
			bottom := isDark(-1, -1) // Matches quiet zone polarity
			if r+1 < total {
				bottom = isDark(r+1, c)
			}
			switch {
			case top && bottom:
				sb.WriteRune('█')
			case top && !bottom:
				sb.WriteRune('▀')
			case !top && bottom:
				sb.WriteRune('▄')
			default:
				sb.WriteRune(' ')
			}
		}
		sb.WriteByte('\n')
	}

	if q.IncludeLabel {
		if q.LGAName != "" {
			fmt.Fprintf(&sb, "└─ %s, %s ─┘\n", q.LGAName, q.StateName)
		} else {
			fmt.Fprintf(&sb, "└─ %s ─┘\n", q.Payload)
		}
	}
	return sb.String()
}

// ASCII renders the QR code using pure 2-character ASCII/Unicode blocks ("██" and "  ").
func (q *QRCode) ASCII() string {
	n := len(q.Modules)
	qz := min(q.QuietZone, 2)
	total := n + 2*qz

	darkCell := "██"
	lightCell := "  "
	if q.InvertASCII {
		darkCell, lightCell = lightCell, darkCell
	}

	var sb strings.Builder
	for r := 0; r < total; r++ {
		for c := 0; c < total; c++ {
			mr := r - qz
			mc := c - qz
			dark := false
			if mr >= 0 && mr < n && mc >= 0 && mc < n {
				dark = q.Modules[mr][mc]
			}
			if dark {
				sb.WriteString(darkCell)
			} else {
				sb.WriteString(lightCell)
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// ============================================================================
// Built-in 5x7 Pixel Bitmap Font for PNG Label Rendering
// ============================================================================

var bitmapFont5x7 = map[byte][7]uint8{
	' ': {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	'-': {0x00, 0x00, 0x00, 0x1F, 0x00, 0x00, 0x00},
	'.': {0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x0C},
	',': {0x00, 0x00, 0x00, 0x00, 0x04, 0x04, 0x08},
	':': {0x00, 0x0C, 0x0C, 0x00, 0x0C, 0x0C, 0x00},
	'/': {0x01, 0x02, 0x04, 0x04, 0x08, 0x10, 0x00},
	'0': {0x0E, 0x11, 0x13, 0x15, 0x19, 0x11, 0x0E},
	'1': {0x04, 0x0C, 0x04, 0x04, 0x04, 0x04, 0x0E},
	'2': {0x0E, 0x11, 0x01, 0x06, 0x08, 0x10, 0x1F},
	'3': {0x1F, 0x02, 0x04, 0x06, 0x01, 0x11, 0x0E},
	'4': {0x02, 0x06, 0x0A, 0x12, 0x1F, 0x02, 0x02},
	'5': {0x1F, 0x10, 0x1E, 0x01, 0x01, 0x11, 0x0E},
	'6': {0x06, 0x08, 0x10, 0x1E, 0x11, 0x11, 0x0E},
	'7': {0x1F, 0x01, 0x02, 0x04, 0x08, 0x08, 0x08},
	'8': {0x0E, 0x11, 0x11, 0x0E, 0x11, 0x11, 0x0E},
	'9': {0x0E, 0x11, 0x11, 0x0F, 0x01, 0x02, 0x0C},
	'A': {0x0E, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11},
	'B': {0x1E, 0x11, 0x11, 0x1E, 0x11, 0x11, 0x1E},
	'C': {0x0E, 0x11, 0x10, 0x10, 0x10, 0x11, 0x0E},
	'D': {0x1C, 0x12, 0x11, 0x11, 0x11, 0x12, 0x1C},
	'E': {0x1F, 0x10, 0x10, 0x1E, 0x10, 0x10, 0x1F},
	'F': {0x1F, 0x10, 0x10, 0x1E, 0x10, 0x10, 0x10},
	'G': {0x0E, 0x11, 0x10, 0x17, 0x11, 0x11, 0x0F},
	'H': {0x11, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11},
	'I': {0x0E, 0x04, 0x04, 0x04, 0x04, 0x04, 0x0E},
	'J': {0x07, 0x02, 0x02, 0x02, 0x02, 0x12, 0x0C},
	'K': {0x11, 0x12, 0x14, 0x18, 0x14, 0x12, 0x11},
	'L': {0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x1F},
	'M': {0x11, 0x1B, 0x15, 0x15, 0x11, 0x11, 0x11},
	'N': {0x11, 0x19, 0x15, 0x13, 0x11, 0x11, 0x11},
	'O': {0x0E, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0E},
	'P': {0x1E, 0x11, 0x11, 0x1E, 0x10, 0x10, 0x10},
	'Q': {0x0E, 0x11, 0x11, 0x11, 0x15, 0x12, 0x0D},
	'R': {0x1E, 0x11, 0x11, 0x1E, 0x14, 0x12, 0x11},
	'S': {0x0F, 0x10, 0x10, 0x0E, 0x01, 0x01, 0x1E},
	'T': {0x1F, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04},
	'U': {0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0E},
	'V': {0x11, 0x11, 0x11, 0x11, 0x11, 0x0A, 0x04},
	'W': {0x11, 0x11, 0x11, 0x15, 0x15, 0x1B, 0x11},
	'X': {0x11, 0x11, 0x0A, 0x04, 0x0A, 0x11, 0x11},
	'Y': {0x11, 0x11, 0x0A, 0x04, 0x04, 0x04, 0x04},
	'Z': {0x1F, 0x01, 0x02, 0x04, 0x08, 0x10, 0x1F},
}

func drawBitmapTextCentered(img *image.Paletted, text string, centerX, startY, scale int, colorIdx uint8) {
	text = strings.ToUpper(text)
	if len(text) == 0 || scale <= 0 {
		return
	}
	charWidth := 5 * scale
	spacing := scale
	totalW := len(text)*charWidth + (len(text)-1)*spacing
	startX := centerX - totalW/2

	bounds := img.Bounds()
	for idx := 0; idx < len(text); idx++ {
		glyph, ok := bitmapFont5x7[text[idx]]
		if !ok {
			glyph = bitmapFont5x7[' ']
		}
		gx := startX + idx*(charWidth+spacing)
		for row := 0; row < 7; row++ {
			rowBits := glyph[row]
			for col := 0; col < 5; col++ {
				if (rowBits>>(4-col))&1 == 0 {
					continue
				}
				px := gx + col*scale
				py := startY + row*scale
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						x, y := px+dx, py+dy
						if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
							img.Pix[y*img.Stride+x] = colorIdx
						}
					}
				}
			}
		}
	}
}

// ============================================================================
// ISO/IEC 18004 QR Code Encoder (Byte Mode, Versions 1-20, Levels L/M/Q/H)
// ============================================================================

type qrVersionSpec struct {
	ecPerBlock int
	b1, d1     int
	b2, d2     int
}

func (s qrVersionSpec) totalData() int {
	return s.b1*s.d1 + s.b2*s.d2
}

// ISO/IEC 18004 Table 9: Error correction block specifications for Versions 1..20.
// Indexed by [version][ecLevelIndex] where L=0, M=1, Q=2, H=3.
var qrVersionSpecs = [21][4]qrVersionSpec{
	{}, // Version 0 (unused)
	// Version 1 (26 total codewords)
	{{7, 1, 19, 0, 0}, {10, 1, 16, 0, 0}, {13, 1, 13, 0, 0}, {17, 1, 9, 0, 0}},
	// Version 2 (44 total codewords)
	{{10, 1, 34, 0, 0}, {16, 1, 28, 0, 0}, {22, 1, 22, 0, 0}, {28, 1, 16, 0, 0}},
	// Version 3 (70 total codewords)
	{{15, 1, 55, 0, 0}, {26, 1, 44, 0, 0}, {18, 2, 17, 0, 0}, {22, 2, 13, 0, 0}},
	// Version 4 (100 total codewords)
	{{20, 1, 80, 0, 0}, {18, 2, 32, 0, 0}, {26, 2, 24, 0, 0}, {16, 4, 9, 0, 0}},
	// Version 5 (134 total codewords)
	{{26, 1, 108, 0, 0}, {24, 2, 43, 0, 0}, {18, 2, 15, 2, 16}, {22, 2, 11, 2, 12}},
	// Version 6 (172 total codewords)
	{{18, 2, 68, 0, 0}, {16, 4, 27, 0, 0}, {24, 4, 19, 0, 0}, {28, 4, 15, 0, 0}},
	// Version 7 (196 total codewords)
	{{20, 2, 78, 0, 0}, {18, 4, 31, 0, 0}, {18, 2, 14, 4, 15}, {26, 4, 13, 1, 14}},
	// Version 8 (242 total codewords)
	{{24, 2, 97, 0, 0}, {22, 2, 38, 2, 39}, {22, 4, 18, 2, 19}, {26, 4, 14, 2, 15}},
	// Version 9 (292 total codewords)
	{{30, 2, 116, 0, 0}, {22, 3, 36, 2, 37}, {20, 4, 16, 4, 17}, {24, 4, 12, 4, 13}},
	// Version 10 (346 total codewords)
	{{18, 2, 68, 2, 69}, {26, 4, 43, 1, 44}, {24, 6, 19, 2, 20}, {28, 6, 15, 2, 16}},
	// Version 11 (404 total codewords)
	{{20, 4, 81, 0, 0}, {30, 1, 50, 4, 51}, {28, 4, 22, 4, 23}, {24, 3, 12, 8, 13}},
	// Version 12 (466 total codewords)
	{{24, 2, 92, 2, 93}, {22, 6, 36, 2, 37}, {26, 4, 20, 6, 21}, {28, 7, 14, 4, 15}},
	// Version 13 (532 total codewords)
	{{26, 4, 107, 0, 0}, {22, 8, 37, 1, 38}, {24, 8, 20, 4, 21}, {22, 12, 11, 4, 12}},
	// Version 14 (581 total codewords)
	{{30, 3, 115, 1, 116}, {24, 4, 40, 5, 41}, {20, 11, 16, 5, 17}, {24, 11, 12, 5, 13}},
	// Version 15 (655 total codewords)
	{{22, 5, 87, 1, 88}, {24, 5, 41, 5, 42}, {30, 5, 24, 7, 25}, {24, 11, 12, 7, 13}},
	// Version 16 (733 total codewords)
	{{24, 5, 98, 1, 99}, {28, 7, 45, 3, 46}, {24, 15, 19, 2, 20}, {30, 3, 15, 13, 16}},
	// Version 17 (815 total codewords)
	{{28, 1, 107, 5, 108}, {28, 10, 46, 1, 47}, {28, 1, 22, 15, 23}, {28, 2, 14, 17, 15}},
	// Version 18 (901 total codewords)
	{{30, 5, 120, 1, 121}, {26, 9, 43, 4, 44}, {28, 17, 22, 1, 23}, {28, 2, 14, 19, 15}},
	// Version 19 (991 total codewords)
	{{28, 3, 113, 4, 114}, {26, 3, 44, 11, 45}, {26, 17, 21, 4, 22}, {26, 9, 13, 16, 14}},
	// Version 20 (1085 total codewords)
	{{28, 3, 107, 5, 108}, {26, 3, 41, 13, 42}, {30, 15, 24, 5, 25}, {28, 15, 15, 10, 16}},
}

// ISO/IEC 18004 Annex E Table E.1: Alignment pattern center coordinates for Versions 1..20.
var qrAlignmentCoords = [21][]int{
	nil,
	nil,
	{6, 18},
	{6, 22},
	{6, 26},
	{6, 30},
	{6, 34},
	{6, 22, 38},
	{6, 24, 42},
	{6, 26, 46},
	{6, 28, 50},
	{6, 30, 54},
	{6, 32, 58},
	{6, 34, 62},
	{6, 26, 46, 66},
	{6, 26, 48, 70},
	{6, 26, 50, 74},
	{6, 30, 54, 78},
	{6, 30, 56, 82},
	{6, 30, 58, 86},
	{6, 34, 62, 90},
}

func ecLevelIndex(lvl QRErrorCorrectionLevel) int {
	switch lvl {
	case QRLevelL:
		return 0
	case QRLevelM:
		return 1
	case QRLevelQ:
		return 2
	case QRLevelH:
		return 3
	default:
		return 1
	}
}

func ecFormatBitsIndicator(lvl QRErrorCorrectionLevel) int {
	// Per ISO/IEC 18004 Table 25: L=01(1), M=00(0), Q=11(3), H=10(2)
	switch lvl {
	case QRLevelL:
		return 1
	case QRLevelM:
		return 0
	case QRLevelQ:
		return 3
	case QRLevelH:
		return 2
	default:
		return 0
	}
}

func encodeQRMatrix(data []byte, level QRErrorCorrectionLevel) ([][]bool, int, error) {
	lvlIdx := ecLevelIndex(level)

	// Select smallest version (1..20) that fits data in Byte mode
	version := 0
	for v := 1; v <= 20; v++ {
		charCountBits := 8
		if v >= 10 {
			charCountBits = 16
		}
		neededBits := 4 + charCountBits + len(data)*8
		capacityBits := qrVersionSpecs[v][lvlIdx].totalData() * 8
		if neededBits <= capacityBits {
			version = v
			break
		}
	}
	if version == 0 {
		return nil, 0, fmt.Errorf("postcode: QR payload length (%d bytes) exceeds maximum supported capacity", len(data))
	}

	spec := qrVersionSpecs[version][lvlIdx]
	dataCodewords := buildDataCodewords(data, version, spec.totalData())
	allCodewords := interleaveWithECC(dataCodewords, spec)

	size := 17 + 4*version
	modules := make([][]bool, size)
	isFunction := make([][]bool, size)
	for r := 0; r < size; r++ {
		modules[r] = make([]bool, size)
		isFunction[r] = make([]bool, size)
	}

	drawFunctionPatterns(modules, isFunction, version, level, 0)
	drawCodewords(modules, isFunction, allCodewords)

	// Evaluate all 8 mask patterns and select the lowest penalty score
	bestMask := 0
	bestScore := math.MaxInt
	for mask := 0; mask < 8; mask++ {
		applyMask(modules, isFunction, mask)
		drawFormatBits(modules, isFunction, level, mask)
		score := computePenaltyScore(modules)
		if score < bestScore {
			bestScore = score
			bestMask = mask
		}
		applyMask(modules, isFunction, mask) // XOR again to restore unmasked state
	}

	applyMask(modules, isFunction, bestMask)
	drawFormatBits(modules, isFunction, level, bestMask)

	return modules, version, nil
}

type bitBuffer struct {
	bits []uint8
}

func (b *bitBuffer) appendBits(val int, numBits int) {
	for i := numBits - 1; i >= 0; i-- {
		b.bits = append(b.bits, uint8((val>>i)&1))
	}
}

func buildDataCodewords(data []byte, version int, totalDataCodewords int) []byte {
	capacityBits := totalDataCodewords * 8
	var bb bitBuffer
	bb.bits = make([]uint8, 0, capacityBits)

	// 1. Byte mode indicator (0100 = 4)
	bb.appendBits(0b0100, 4)

	// 2. Character count indicator (8 bits for V1-9, 16 bits for V10-40)
	ccBits := 8
	if version >= 10 {
		ccBits = 16
	}
	bb.appendBits(len(data), ccBits)

	// 3. Payload bytes
	for _, b := range data {
		bb.appendBits(int(b), 8)
	}

	// 4. Terminator (up to 4 zero bits)
	termBits := min(4, capacityBits-len(bb.bits))
	bb.appendBits(0, termBits)

	// 5. Pad with zero bits to byte boundary
	if rem := len(bb.bits) % 8; rem != 0 {
		bb.appendBits(0, 8-rem)
	}

	// 6. Convert bits to bytes and pad with alternating 0xEC, 0x11
	codewords := make([]byte, totalDataCodewords)
	numBytes := len(bb.bits) / 8
	for i := 0; i < numBytes; i++ {
		var cw byte
		for j := 0; j < 8; j++ {
			cw = (cw << 1) | bb.bits[i*8+j]
		}
		codewords[i] = cw
	}

	padBytes := [2]byte{0xEC, 0x11}
	for i := numBytes; i < totalDataCodewords; i++ {
		codewords[i] = padBytes[(i-numBytes)%2]
	}

	return codewords
}

// ============================================================================
// Reed-Solomon GF(256) Error Correction & Block Interleaving
// ============================================================================

var (
	gfExp [256]byte
	gfLog [256]byte
)

func init() {
	x := 1
	for i := 0; i < 255; i++ {
		gfExp[i] = byte(x)
		gfLog[x] = byte(i)
		x <<= 1
		if x&0x100 != 0 {
			x ^= 0x11D // x^8 + x^4 + x^3 + x^2 + 1
		}
	}
	gfExp[255] = gfExp[0]
}

func gfMul(x, y byte) byte {
	if x == 0 || y == 0 {
		return 0
	}
	return gfExp[(int(gfLog[x])+int(gfLog[y]))%255]
}

func rsGeneratorDivisor(degree int) []byte {
	// Computes coefficients of (x - alpha^0)(x - alpha^1)...(x - alpha^(degree-1))
	// excluding the leading x^degree term, ordered from x^(degree-1) down to x^0.
	result := make([]byte, degree)
	result[degree-1] = 1
	root := byte(1)
	for i := 0; i < degree; i++ {
		for j := 0; j < degree; j++ {
			result[j] = gfMul(result[j], root)
			if j+1 < degree {
				result[j] ^= result[j+1]
			}
		}
		root = gfMul(root, 0x02)
	}
	return result
}

func rsComputeRemainder(data []byte, divisor []byte) []byte {
	result := make([]byte, len(divisor))
	for _, b := range data {
		factor := b ^ result[0]
		copy(result, result[1:])
		result[len(result)-1] = 0
		for i, coef := range divisor {
			result[i] ^= gfMul(coef, factor)
		}
	}
	return result
}

func interleaveWithECC(data []byte, spec qrVersionSpec) []byte {
	numBlocks := spec.b1 + spec.b2
	dataBlocks := make([][]byte, numBlocks)
	eccBlocks := make([][]byte, numBlocks)
	divisor := rsGeneratorDivisor(spec.ecPerBlock)

	offset := 0
	for i := 0; i < numBlocks; i++ {
		blockLen := spec.d1
		if i >= spec.b1 {
			blockLen = spec.d2
		}
		blk := data[offset : offset+blockLen]
		offset += blockLen
		dataBlocks[i] = blk
		eccBlocks[i] = rsComputeRemainder(blk, divisor)
	}

	maxDataLen := max(spec.d1, spec.d2)
	totalCodewords := spec.totalData() + numBlocks*spec.ecPerBlock
	out := make([]byte, 0, totalCodewords)

	// Interleave data codewords
	for col := 0; col < maxDataLen; col++ {
		for b := 0; b < numBlocks; b++ {
			if col < len(dataBlocks[b]) {
				out = append(out, dataBlocks[b][col])
			}
		}
	}

	// Interleave ECC codewords
	for col := 0; col < spec.ecPerBlock; col++ {
		for b := 0; b < numBlocks; b++ {
			out = append(out, eccBlocks[b][col])
		}
	}

	return out
}

// ============================================================================
// Function Patterns, Codeword Zigzag Placement & Masking
// ============================================================================

func setFunctionModule(modules, isFunction [][]bool, r, c int, dark bool) {
	modules[r][c] = dark
	isFunction[r][c] = true
}

func drawFunctionPatterns(modules, isFunction [][]bool, version int, level QRErrorCorrectionLevel, mask int) {
	size := len(modules)

	// 1. Timing patterns (row 6 and col 6)
	for i := 0; i < size; i++ {
		setFunctionModule(modules, isFunction, 6, i, i%2 == 0)
		setFunctionModule(modules, isFunction, i, 6, i%2 == 0)
	}

	// 2. Three 7x7 finder patterns + 1-module white separators (centered at (3,3), (3,size-4), (size-4,3))
	drawFinderPattern(modules, isFunction, 3, 3)
	drawFinderPattern(modules, isFunction, 3, size-4)
	drawFinderPattern(modules, isFunction, size-4, 3)

	// 3. Alignment patterns (5x5)
	coords := qrAlignmentCoords[version]
	for i, r := range coords {
		for j, c := range coords {
			// Skip the 3 corners that overlap finder patterns
			if (i == 0 && j == 0) || (i == 0 && j == len(coords)-1) || (i == len(coords)-1 && j == 0) {
				continue
			}
			drawAlignmentPattern(modules, isFunction, r, c)
		}
	}

	// 4. Format information (reserves format modules + sets dark module)
	drawFormatBits(modules, isFunction, level, mask)

	// 5. Version information (Version >= 7)
	if version >= 7 {
		drawVersionBits(modules, isFunction, version)
	}
}

func drawFinderPattern(modules, isFunction [][]bool, centerR, centerC int) {
	size := len(modules)
	for dy := -4; dy <= 4; dy++ {
		for dx := -4; dx <= 4; dx++ {
			r, c := centerR+dy, centerC+dx
			if r < 0 || r >= size || c < 0 || c >= size {
				continue
			}
			dist := max(absInt(dx), absInt(dy))
			dark := dist != 2 && dist != 4
			setFunctionModule(modules, isFunction, r, c, dark)
		}
	}
}

func drawAlignmentPattern(modules, isFunction [][]bool, centerR, centerC int) {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			dist := max(absInt(dx), absInt(dy))
			setFunctionModule(modules, isFunction, centerR+dy, centerC+dx, dist != 1)
		}
	}
}

func drawFormatBits(modules, isFunction [][]bool, level QRErrorCorrectionLevel, mask int) {
	size := len(modules)
	data := (ecFormatBitsIndicator(level) << 3) | mask
	rem := data
	for i := 0; i < 10; i++ {
		rem = (rem << 1) ^ ((rem >> 9) * 0x537)
	}
	bits := ((data << 10) | rem) ^ 0x5412

	getBit := func(idx int) bool {
		return ((bits >> idx) & 1) != 0
	}

	// First copy around top-left finder
	for i := 0; i <= 5; i++ {
		setFunctionModule(modules, isFunction, i, 8, getBit(i))
	}
	setFunctionModule(modules, isFunction, 7, 8, getBit(6))
	setFunctionModule(modules, isFunction, 8, 8, getBit(7))
	setFunctionModule(modules, isFunction, 8, 7, getBit(8))
	for i := 9; i < 15; i++ {
		setFunctionModule(modules, isFunction, 8, 14-i, getBit(i))
	}

	// Second copy split between top-right and bottom-left finders
	for i := 0; i < 8; i++ {
		setFunctionModule(modules, isFunction, 8, size-1-i, getBit(i))
	}
	for i := 8; i < 15; i++ {
		setFunctionModule(modules, isFunction, size-15+i, 8, getBit(i))
	}

	// Always-dark module at (row = size-8, col = 8)
	setFunctionModule(modules, isFunction, size-8, 8, true)
}

func drawVersionBits(modules, isFunction [][]bool, version int) {
	size := len(modules)
	rem := version
	for i := 0; i < 12; i++ {
		rem = (rem << 1) ^ ((rem >> 11) * 0x1F25)
	}
	bits := (version << 12) | rem

	for i := 0; i < 18; i++ {
		bit := ((bits >> i) & 1) != 0
		a := size - 11 + (i % 3)
		b := i / 3
		// Top-right 6x3 block (rows 0..5, cols size-11..size-9)
		setFunctionModule(modules, isFunction, b, a, bit)
		// Bottom-left 3x6 block (rows size-11..size-9, cols 0..5)
		setFunctionModule(modules, isFunction, a, b, bit)
	}
}

func drawCodewords(modules, isFunction [][]bool, codewords []byte) {
	size := len(modules)
	bitIdx := 0
	totalBits := len(codewords) * 8
	upward := true

	for right := size - 1; right >= 1; right -= 2 {
		if right == 6 {
			right = 5 // Skip vertical timing pattern column
		}
		for step := 0; step < size; step++ {
			r := step
			if upward {
				r = size - 1 - step
			}
			for j := 0; j < 2; j++ {
				c := right - j
				if !isFunction[r][c] {
					dark := false
					if bitIdx < totalBits {
						dark = ((codewords[bitIdx>>3] >> (7 - (bitIdx & 7))) & 1) != 0
						bitIdx++
					}
					modules[r][c] = dark
				}
			}
		}
		upward = !upward
	}
}

func maskBit(mask, r, c int) bool {
	switch mask {
	case 0:
		return (r+c)%2 == 0
	case 1:
		return r%2 == 0
	case 2:
		return c%3 == 0
	case 3:
		return (r+c)%3 == 0
	case 4:
		return (r/2+c/3)%2 == 0
	case 5:
		return (r*c)%2+(r*c)%3 == 0
	case 6:
		return ((r*c)%2+(r*c)%3)%2 == 0
	case 7:
		return ((r+c)%2+(r*c)%3)%2 == 0
	default:
		return false
	}
}

func applyMask(modules, isFunction [][]bool, mask int) {
	size := len(modules)
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			if !isFunction[r][c] && maskBit(mask, r, c) {
				modules[r][c] = !modules[r][c]
			}
		}
	}
}

func computePenaltyScore(modules [][]bool) int {
	size := len(modules)
	score := 0

	// Rule 1: 5+ consecutive same-color modules in rows and columns
	for r := 0; r < size; r++ {
		runColor := modules[r][0]
		runLen := 1
		for c := 1; c < size; c++ {
			if modules[r][c] == runColor {
				runLen++
				if runLen == 5 {
					score += 3
				} else if runLen > 5 {
					score++
				}
			} else {
				runColor = modules[r][c]
				runLen = 1
			}
		}
	}
	for c := 0; c < size; c++ {
		runColor := modules[0][c]
		runLen := 1
		for r := 1; r < size; r++ {
			if modules[r][c] == runColor {
				runLen++
				if runLen == 5 {
					score += 3
				} else if runLen > 5 {
					score++
				}
			} else {
				runColor = modules[r][c]
				runLen = 1
			}
		}
	}

	// Rule 2: 2x2 blocks of identical color
	for r := 0; r < size-1; r++ {
		for c := 0; c < size-1; c++ {
			val := modules[r][c]
			if modules[r][c+1] == val && modules[r+1][c] == val && modules[r+1][c+1] == val {
				score += 3
			}
		}
	}

	// Rule 3: 1:1:3:1:1 finder-like patterns preceded or followed by 4 light modules
	for r := 0; r < size; r++ {
		bits := 0
		for c := 0; c < size; c++ {
			bits = ((bits << 1) & 0x7FF)
			if modules[r][c] {
				bits |= 1
			}
			if c >= 10 && (bits == 0x05D || bits == 0x5D0) {
				score += 40
			}
		}
	}
	for c := 0; c < size; c++ {
		bits := 0
		for r := 0; r < size; r++ {
			bits = ((bits << 1) & 0x7FF)
			if modules[r][c] {
				bits |= 1
			}
			if r >= 10 && (bits == 0x05D || bits == 0x5D0) {
				score += 40
			}
		}
	}

	// Rule 4: Proportion of dark modules relative to 50%
	dark := 0
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			if modules[r][c] {
				dark++
			}
		}
	}
	total := size * size
	k := (absInt(dark*20-total*10) + total - 1) / total
	if k > 1 {
		score += (k - 1) * 10
	}

	return score
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
