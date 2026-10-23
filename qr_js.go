//go:build js && wasm

package postcode

import "io"

// WritePNG writes a standards-compliant 8-bit indexed PNG image to w without
// importing image, image/png, compress/zlib, or os.
func (q *QRCode) WritePNG(w io.Writer) error {
	pix, width, height, pal := q.renderPalettedBuffer()

	// Raw scanlines: each row is 1 filter byte (0x00 = None) + width palette index bytes
	rowLen := 1 + width
	rawLen := height * rowLen
	raw := make([]byte, rawLen)
	for y := 0; y < height; y++ {
		dstOff := y * rowLen
		raw[dstOff] = 0 // Filter type 0 (None)
		copy(raw[dstOff+1:dstOff+rowLen], pix[y*width:(y+1)*width])
	}

	// Build ZLIB stream using uncompressed DEFLATE blocks (max 65535 bytes per block)
	const maxBlock = 65535
	numBlocks := (rawLen + maxBlock - 1) / maxBlock
	if numBlocks == 0 {
		numBlocks = 1
	}
	zlibData := make([]byte, 0, 2+numBlocks*5+rawLen+4)
	zlibData = append(zlibData, 0x78, 0x01) // ZLIB header (CM=8, CINFO=7, no preset dict)

	rem := raw
	for i := 0; i < numBlocks; i++ {
		chunkLen := len(rem)
		if chunkLen > maxBlock {
			chunkLen = maxBlock
		}
		final := byte(0x00)
		if i == numBlocks-1 {
			final = 0x01
		}
		nLen := ^uint16(chunkLen)
		zlibData = append(zlibData,
			final,
			byte(chunkLen), byte(chunkLen>>8),
			byte(nLen), byte(nLen>>8),
		)
		zlibData = append(zlibData, rem[:chunkLen]...)
		rem = rem[chunkLen:]
	}

	adler := pngAdler32(raw)
	zlibData = append(zlibData, byte(adler>>24), byte(adler>>16), byte(adler>>8), byte(adler))

	// Assemble PNG file: signature + IHDR + PLTE + IDAT + IEND
	out := make([]byte, 0, 64+len(zlibData))
	out = append(out, 0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n')

	ihdr := [13]byte{
		byte(width >> 24), byte(width >> 16), byte(width >> 8), byte(width),
		byte(height >> 24), byte(height >> 16), byte(height >> 8), byte(height),
		8, // bit depth: 8
		3, // color type: 3 (Indexed-color)
		0, // compression method: 0
		0, // filter method: 0
		0, // interlace method: 0
	}
	out = appendPNGChunk(out, "IHDR", ihdr[:])

	plte := [9]byte{
		pal[0][0], pal[0][1], pal[0][2],
		pal[1][0], pal[1][1], pal[1][2],
		pal[2][0], pal[2][1], pal[2][2],
	}
	out = appendPNGChunk(out, "PLTE", plte[:])
	out = appendPNGChunk(out, "IDAT", zlibData)
	out = appendPNGChunk(out, "IEND", nil)

	_, err := w.Write(out)
	return err
}

func appendPNGChunk(dst []byte, chunkType string, data []byte) []byte {
	n := uint32(len(data))
	dst = append(dst, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	typeStart := len(dst)
	dst = append(dst, chunkType...)
	dst = append(dst, data...)
	crc := pngCRC32(dst[typeStart:])
	dst = append(dst, byte(crc>>24), byte(crc>>16), byte(crc>>8), byte(crc))
	return dst
}

func pngCRC32(data []byte) uint32 {
	crc := uint32(0xFFFFFFFF)
	for _, b := range data {
		crc ^= uint32(b)
		for k := 0; k < 8; k++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xEDB88320
			} else {
				crc >>= 1
			}
		}
	}
	return ^crc
}

func pngAdler32(data []byte) uint32 {
	s1 := uint32(1)
	s2 := uint32(0)
	for _, b := range data {
		s1 = (s1 + uint32(b)) % 65521
		s2 = (s2 + s1) % 65521
	}
	return (s2 << 16) | s1
}
