package watermark

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
)

// Embedded reference watermark images.
// These are screenshots of the Gemini watermark applied to a black background.
// They are used to extract alpha values for the reverse blending calculation.
//
// The images are embedded at compile time using Go's embed directive,
// making the binary self-contained with no external file dependencies.
var (
	//go:embed assets/bg_48.png
	bg48PNG []byte

	//go:embed assets/bg_96.png
	bg96PNG []byte
)

// LoadReferenceImage loads and decodes the embedded reference watermark image
// for the specified size.
//
// Parameters:
//   - size: The watermark size to load (48 or 96). Any other value defaults to 48.
//
// Returns:
//   - image.Image: The decoded PNG image
//   - error: An error if PNG decoding fails
//
// The reference images contain the Gemini watermark rendered on a black background.
// Since the watermark is white with varying alpha, each pixel's brightness
// directly corresponds to its alpha value in the original watermark.
func LoadReferenceImage(size int) (image.Image, error) {
	var data []byte

	switch size {
	case 96:
		data = bg96PNG
	case 48:
		fallthrough
	default:
		data = bg48PNG
	}

	return png.Decode(bytes.NewReader(data))
}
