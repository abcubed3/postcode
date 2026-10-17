package postcode

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// decodeQRMatrixForTest is an independent ISO/IEC 18004 matrix decoder used to verify
// that generated QR matrices have valid format info, zero Reed-Solomon syndromes,
// and decode back to the exact input byte payload.
func decodeQRMatrixForTest(t *testing.T, modules [][]bool, version int, expectedLevel QRErrorCorrectionLevel) string {
	t.Helper()
	size := len(modules)
	expectedSize := 17 + 4*version
	if size != expectedSize {
		t.Fatalf("matrix size = %d, want %d for version %d", size, expectedSize, version)
	}

	// 1. Read format bits from copy 1 (around top-left finder)
	bits1 := 0
	for i := 0; i <= 5; i++ {
		if modules[i][8] {
			bits1 |= 1 << i
		}
	}
	if modules[7][8] {
		bits1 |= 1 << 6
	}
	if modules[8][8] {
		bits1 |= 1 << 7
	}
	if modules[8][7] {
		bits1 |= 1 << 8
	}
	for i := 9; i < 15; i++ {
		if modules[8][14-i] {
			bits1 |= 1 << i
		}
	}

	// Read format bits from copy 2 (top-right and bottom-left)
	bits2 := 0
	for i := 0; i < 8; i++ {
		if modules[8][size-1-i] {
			bits2 |= 1 << i
		}
	}
	for i := 8; i < 15; i++ {
		if modules[size-15+i][8] {
			bits2 |= 1 << i
		}
	}

	if bits1 != bits2 {
		t.Fatalf("format bits mismatch between copy 1 (%015b) and copy 2 (%015b)", bits1, bits2)
	}

	// Verify always-dark module
	if !modules[size-8][8] {
		t.Fatalf("expected dark module at (%d, 8) to be true", size-8)
	}

	unmaskedFormat := bits1 ^ 0x5412
	formatData := unmaskedFormat >> 10
	mask := formatData & 0x7
	ecInd := (formatData >> 3) & 0x3
	if ecInd != ecFormatBitsIndicator(expectedLevel) {
		t.Fatalf("decoded EC indicator = %d, want %d", ecInd, ecFormatBitsIndicator(expectedLevel))
	}

	// 2. Reconstruct function module map and unmask a copy of the matrix
	dummy := make([][]bool, size)
	isFunction := make([][]bool, size)
	for r := range size {
		dummy[r] = make([]bool, size)
		isFunction[r] = make([]bool, size)
	}
	drawFunctionPatterns(dummy, isFunction, version, expectedLevel, mask)

	unmasked := make([][]bool, size)
	for r := range size {
		unmasked[r] = make([]bool, size)
		copy(unmasked[r], modules[r])
	}
	applyMask(unmasked, isFunction, mask)

	// 3. Extract raw bits in zigzag order
	var rawBits []uint8
	upward := true
	for right := size - 1; right >= 1; right -= 2 {
		if right == 6 {
			right = 5
		}
		for step := range size {
			r := step
			if upward {
				r = size - 1 - step
			}
			for j := range 2 {
				c := right - j
				if !isFunction[r][c] {
					if unmasked[r][c] {
						rawBits = append(rawBits, 1)
					} else {
						rawBits = append(rawBits, 0)
					}
				}
			}
		}
		upward = !upward
	}

	spec := qrVersionSpecs[version][ecLevelIndex(expectedLevel)]
	numBlocks := spec.b1 + spec.b2
	totalCodewords := spec.totalData() + numBlocks*spec.ecPerBlock
	if len(rawBits) < totalCodewords*8 {
		t.Fatalf("insufficient raw bits: got %d, need at least %d", len(rawBits), totalCodewords*8)
	}

	codewords := make([]byte, totalCodewords)
	for i := range totalCodewords {
		var b byte
		for j := range 8 {
			b = (b << 1) | rawBits[i*8+j]
		}
		codewords[i] = b
	}

	// 4. De-interleave data and ECC blocks, and verify Reed-Solomon remainder == 0
	dataBlocks := make([][]byte, numBlocks)
	eccBlocks := make([][]byte, numBlocks)
	for i := range numBlocks {
		blockLen := spec.d1
		if i >= spec.b1 {
			blockLen = spec.d2
		}
		dataBlocks[i] = make([]byte, blockLen)
		eccBlocks[i] = make([]byte, spec.ecPerBlock)
	}

	idx := 0
	maxDataLen := max(spec.d1, spec.d2)
	for col := range maxDataLen {
		for b := range numBlocks {
			if col < len(dataBlocks[b]) {
				dataBlocks[b][col] = codewords[idx]
				idx++
			}
		}
	}
	for col := range spec.ecPerBlock {
		for b := range numBlocks {
			eccBlocks[b][col] = codewords[idx]
			idx++
		}
	}

	divisor := rsGeneratorDivisor(spec.ecPerBlock)
	var dataStream []byte
	for b := range numBlocks {
		rem := rsComputeRemainder(dataBlocks[b], divisor)
		if !bytes.Equal(rem, eccBlocks[b]) {
			t.Fatalf("block %d Reed-Solomon ECC mismatch: got %x, want %x", b, eccBlocks[b], rem)
		}
		dataStream = append(dataStream, dataBlocks[b]...)
	}

	// 5. Decode Byte Mode bitstream
	var streamBits []uint8
	for _, cw := range dataStream {
		for j := 7; j >= 0; j-- {
			streamBits = append(streamBits, (cw>>j)&1)
		}
	}

	readBits := func(pos, n int) int {
		v := 0
		for i := range n {
			v = (v << 1) | int(streamBits[pos+i])
		}
		return v
	}

	mode := readBits(0, 4)
	if mode != 0b0100 {
		t.Fatalf("expected Byte Mode (4), got %d", mode)
	}
	ccBits := 8
	if version >= 10 {
		ccBits = 16
	}
	charCount := readBits(4, ccBits)
	bitOffset := 4 + ccBits
	decoded := make([]byte, charCount)
	for i := range charCount {
		decoded[i] = byte(readBits(bitOffset+i*8, 8))
	}

	return string(decoded)
}

func TestQRVersionSpecsInvariants(t *testing.T) {
	levels := []QRErrorCorrectionLevel{QRLevelL, QRLevelM, QRLevelQ, QRLevelH}
	for v := 1; v <= 20; v++ {
		size := 17 + 4*v
		dummy := make([][]bool, size)
		isFunc := make([][]bool, size)
		for r := range size {
			dummy[r] = make([]bool, size)
			isFunc[r] = make([]bool, size)
		}
		drawFunctionPatterns(dummy, isFunc, v, QRLevelM, 0)

		availBits := 0
		for r := range size {
			for c := range size {
				if !isFunc[r][c] {
					availBits++
				}
			}
		}
		expectedCodewords := availBits / 8

		for _, lvl := range levels {
			spec := qrVersionSpecs[v][ecLevelIndex(lvl)]
			totalCW := spec.totalData() + (spec.b1+spec.b2)*spec.ecPerBlock
			if totalCW != expectedCodewords {
				t.Errorf("version %d level %s: spec total codewords = %d, matrix available codewords = %d",
					v, lvl, totalCW, expectedCodewords)
			}
			if spec.b2 > 0 && spec.d2 != spec.d1+1 {
				t.Errorf("version %d level %s: d2 (%d) != d1+1 (%d)", v, lvl, spec.d2, spec.d1+1)
			}
		}
	}
}

func TestGenerateQRRoundTripPayloads(t *testing.T) {
	payloads := []QRPayloadType{
		QRPayloadPostcode,
		QRPayloadCompact,
		QRPayloadSpaced,
		QRPayloadGoogleMaps,
		QRPayloadDirections,
		QRPayloadAppleMaps,
		QRPayloadOSM,
		QRPayloadGeoURI,
		QRPayloadJSON,
		QRPayloadText,
	}
	levels := []QRErrorCorrectionLevel{QRLevelL, QRLevelM, QRLevelQ, QRLevelH}

	for _, pt := range payloads {
		for _, lvl := range levels {
			t.Run(string(pt)+"_"+string(lvl), func(t *testing.T) {
				qr, err := GenerateQR("EK 01 A03 FK 01", QROptions{
					Payload: pt,
					Level:   lvl,
				})
				if err != nil {
					t.Fatalf("GenerateQR(%s, %s) error: %v", pt, lvl, err)
				}
				if qr.Postcode != "EK-01-A03-FK-01" {
					t.Errorf("Postcode = %q, want EK-01-A03-FK-01", qr.Postcode)
				}
				if qr.Payload == "" {
					t.Fatalf("expected non-empty payload for %s", pt)
				}

				decoded := decodeQRMatrixForTest(t, qr.Modules, qr.Version, qr.Level)
				if decoded != qr.Payload {
					t.Errorf("decoded QR payload = %q, want %q", decoded, qr.Payload)
				}
			})
		}
	}
}

func TestPostcodeQRRenderFormats(t *testing.T) {
	p := MustParse("LA-11-W06-TC-10")

	// 1. Default QR (canonical postcode payload)
	qr, err := p.QR()
	if err != nil {
		t.Fatalf("p.QR() failed: %v", err)
	}
	if qr.Payload != "LA-11-W06-TC-10" {
		t.Errorf("default payload = %q, want LA-11-W06-TC-10", qr.Payload)
	}
	if qr.StateName != "Lagos" {
		t.Errorf("StateName = %q, want Lagos", qr.StateName)
	}

	// 2. Matrix SVG
	svg, err := p.QRSVG()
	if err != nil {
		t.Fatalf("p.QRSVG() failed: %v", err)
	}
	if !strings.HasPrefix(svg, "<svg ") || !strings.HasSuffix(svg, "</svg>") {
		t.Errorf("expected valid SVG wrapper, got %q", svg[:min(len(svg), 60)])
	}
	if !strings.Contains(svg, "LA-11-W06-TC-10") {
		t.Errorf("expected SVG to include postcode title")
	}

	// 3. Printable Label SVG & HTML
	labelQR, err := GenerateQR("LA-11-W06-TC-10", QROptions{
		IncludeLabel:    true,
		Title:           "NIPOST DELIVERY PARCEL",
		Subtitle:        "10 Obafemi Awolowo Way, Ikeja",
		ForegroundColor: "#00573F",
	})
	if err != nil {
		t.Fatalf("GenerateQR with label failed: %v", err)
	}
	labelSVG := labelQR.SVG()
	if !strings.Contains(labelSVG, "NIPOST DELIVERY PARCEL") || !strings.Contains(labelSVG, "LA-11-W06-TC-10") {
		t.Errorf("expected LabelSVG to contain custom title and postcode")
	}
	if !strings.Contains(labelSVG, "10 Obafemi Awolowo Way, Ikeja") {
		t.Errorf("expected LabelSVG to contain subtitle")
	}

	htmlDoc := labelQR.PrintableHTML()
	if !strings.Contains(htmlDoc, "<!DOCTYPE html>") || !strings.Contains(htmlDoc, "window.print()") {
		t.Errorf("expected PrintableHTML to contain print button and HTML doctype")
	}

	// 4. PNG (both standard and with printable label)
	pngBytes, err := p.QRPNG()
	if err != nil {
		t.Fatalf("p.QRPNG() failed: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("png.Decode failed on standard QR PNG: %v", err)
	}
	if img.Bounds().Dx() != img.Bounds().Dy() {
		t.Errorf("expected square PNG for standard QR, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}

	labelPNGBytes, err := labelQR.PNG()
	if err != nil {
		t.Fatalf("labelQR.PNG() failed: %v", err)
	}
	labelImg, err := png.Decode(bytes.NewReader(labelPNGBytes))
	if err != nil {
		t.Fatalf("png.Decode failed on label QR PNG: %v", err)
	}
	if labelImg.Bounds().Dy() <= labelImg.Bounds().Dx() {
		t.Errorf("expected label PNG height (%d) > width (%d)", labelImg.Bounds().Dy(), labelImg.Bounds().Dx())
	}

	// 5. DataURI and SVGDataURI
	dataURI, err := qr.DataURI()
	if err != nil {
		t.Fatalf("qr.DataURI() failed: %v", err)
	}
	if !strings.HasPrefix(dataURI, "data:image/png;base64,") {
		t.Errorf("unexpected DataURI prefix: %q", dataURI[:30])
	}
	rawPNG, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURI, "data:image/png;base64,"))
	if err != nil || len(rawPNG) == 0 {
		t.Errorf("failed to base64 decode DataURI: %v", err)
	}

	svgURI := qr.SVGDataURI()
	if !strings.HasPrefix(svgURI, "data:image/svg+xml;base64,") {
		t.Errorf("unexpected SVGDataURI prefix: %q", svgURI[:30])
	}

	// 6. Terminal & ASCII
	term, err := p.QRTerminal()
	if err != nil {
		t.Fatalf("p.QRTerminal() failed: %v", err)
	}
	if !strings.ContainsAny(term, "▀▄█") {
		t.Errorf("expected Unicode half-blocks in Terminal output")
	}
	ascii := labelQR.ASCII()
	if !strings.Contains(ascii, "██") {
		t.Errorf("expected block characters in ASCII output")
	}
}

type qrMockRoundTripper func(*http.Request) (*http.Response, error)

func (f qrMockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestClientGenerateQR(t *testing.T) {
	mockRT := qrMockRoundTripper(func(r *http.Request) (*http.Response, error) {
		body, _ := json.Marshal(LookupResponse{
			Postcode: "LA-01-A01-AA-99",
			Valid:    true,
			RecentHouseAddress: &RecentHouseAddress{
				Address: "42 Marina Road, Lagos",
			},
			PointGeometry: &PointGeometry{
				Type:        "Point",
				Coordinates: []float64{3.3903, 6.4541},
			},
		})
		rec := httptest.NewRecorder()
		rec.Header().Set("Content-Type", "application/json")
		rec.WriteHeader(http.StatusOK)
		_, _ = rec.Write(body)
		return rec.Result(), nil
	})

	client, err := NewClient(WithTransport(mockRT), WithAPIKey("test_key"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	qr, err := client.GenerateQR(context.Background(), "LA-01-A01-AA-99", QROptions{
		Payload:      QRPayloadGoogleMaps,
		IncludeLabel: true,
	})
	if err != nil {
		t.Fatalf("client.GenerateQR failed: %v", err)
	}
	if qr.Address != "42 Marina Road, Lagos" {
		t.Errorf("expected enriched address '42 Marina Road, Lagos', got %q", qr.Address)
	}
	if !strings.Contains(qr.LabelSVG(), "42 Marina Road, Lagos") {
		t.Errorf("expected LabelSVG to render enriched street address")
	}
}

func TestGenerateQRErrors(t *testing.T) {
	if _, err := GenerateQR("INVALID"); err == nil {
		t.Errorf("expected error for invalid postcode")
	}
	if _, err := GenerateQR("ZZ-01-A03-FK-01"); err == nil {
		t.Errorf("expected error for unknown state ZZ")
	}
	if _, err := GenerateQR("EK-01-A03-FK-01", QROptions{Payload: "bogus"}); err == nil {
		t.Errorf("expected error for unsupported QR payload type")
	}
	if _, err := GenerateQR("EK-01-A03-FK-01", QROptions{Level: "Z"}); err == nil {
		t.Errorf("expected error for invalid QR error correction level")
	}
	var zero Postcode
	if _, err := zero.QR(); err == nil {
		t.Errorf("expected error for zero Postcode")
	}
}

func TestQRCode_SaveHelpers(t *testing.T) {
	qr, err := GenerateQR("EK-01-A03-FK-01", QROptions{IncludeLabel: true})
	if err != nil {
		t.Fatalf("GenerateQR failed: %v", err)
	}

	tmpDir := t.TempDir()
	svgPath := tmpDir + "/test.svg"
	pngPath := tmpDir + "/test.png"
	htmlPath := tmpDir + "/test.html"

	if err := qr.SaveSVG(svgPath); err != nil {
		t.Fatalf("SaveSVG failed: %v", err)
	}
	if b, err := os.ReadFile(svgPath); err != nil || len(b) == 0 {
		t.Fatalf("expected non-empty saved SVG file")
	}

	if err := qr.SavePNG(pngPath); err != nil {
		t.Fatalf("SavePNG failed: %v", err)
	}
	if b, err := os.ReadFile(pngPath); err != nil || len(b) == 0 {
		t.Fatalf("expected non-empty saved PNG file")
	}

	if err := qr.SaveHTML(htmlPath); err != nil {
		t.Fatalf("SaveHTML failed: %v", err)
	}
	if b, err := os.ReadFile(htmlPath); err != nil || len(b) == 0 {
		t.Fatalf("expected non-empty saved HTML file")
	}
}
