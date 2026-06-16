package image

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
)

// JPEGBytesToPNG перекодирует байты JPEG в PNG.
func JPEGBytesToPNG(jpegBytes []byte) ([]byte, error) {
	img, err := jpeg.Decode(bytes.NewReader(jpegBytes))
	if err != nil {
		return nil, fmt.Errorf("jpeg decode: %w", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("png encode: %w", err)
	}
	return buf.Bytes(), nil
}
