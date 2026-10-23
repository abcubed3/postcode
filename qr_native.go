//go:build !js

package postcode

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
)

// Image renders the QRCode as a standard Go image.Image.
// When IncludeLabel is true, a footer banner with the formatted postcode and State/LGA is drawn below the QR matrix.
func (q *QRCode) Image() image.Image {
	pix, w, h, pal := q.renderPalettedBuffer()
	palette := color.Palette{
		color.RGBA{R: pal[0][0], G: pal[0][1], B: pal[0][2], A: 255},
		color.RGBA{R: pal[1][0], G: pal[1][1], B: pal[1][2], A: 255},
		color.RGBA{R: pal[2][0], G: pal[2][1], B: pal[2][2], A: 255},
	}
	return &image.Paletted{
		Pix:     pix,
		Stride:  w,
		Rect:    image.Rect(0, 0, w, h),
		Palette: palette,
	}
}

// WritePNG writes the PNG-encoded QR code to w.
func (q *QRCode) WritePNG(w io.Writer) error {
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	return enc.Encode(w, q.Image())
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
