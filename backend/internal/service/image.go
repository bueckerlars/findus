package service

import (
	"bytes"
	_ "image/jpeg"
	_ "image/png"
	"io"

	nativewebp "github.com/HugoSmits86/nativewebp"
	"github.com/disintegration/imaging"
)

const maxImageWidth = 1200

// EncodeWebPFromImage reads an image stream, resizes to max width, and encodes WebP.
func EncodeWebPFromImage(r io.Reader) ([]byte, error) {
	img, err := imaging.Decode(r, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	if img.Bounds().Dx() > maxImageWidth {
		img = imaging.Resize(img, maxImageWidth, 0, imaging.Lanczos)
	}
	var buf bytes.Buffer
	if err := nativewebp.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
