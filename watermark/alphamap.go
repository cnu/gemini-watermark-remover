package watermark

import (
	"image"
)

// CalculateAlphaMap extracts alpha (transparency) values from a reference watermark image.
//
// The reference image should be a screenshot of the Gemini watermark applied to a
// pure black background. Since the watermark is white with varying transparency,
// the brightness of each pixel directly corresponds to its alpha value:
//
//   - Pure black (0, 0, 0) = fully transparent (alpha = 0.0)
//   - Pure white (255, 255, 255) = fully opaque (alpha = 1.0)
//   - Gray values = partial transparency
//
// We use max(R, G, B) instead of averaging to handle any slight color variations
// in the captured watermark.
//
// The returned slice contains one float32 per pixel in row-major order,
// with values normalized to the range [0.0, 1.0].
func CalculateAlphaMap(img image.Image) []float32 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Allocate alpha map: one float32 per pixel
	alphaMap := make([]float32, width*height)

	// Process each pixel in row-major order
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Get pixel color values.
			// RGBA() returns values in [0, 65535] range for 16-bit precision.
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()

			// Convert from 16-bit [0, 65535] to 8-bit [0, 255] by shifting right 8 bits.
			// This matches the precision used in the watermark removal algorithm.
			r8 := r >> 8
			g8 := g >> 8
			b8 := b >> 8

			// Find the maximum channel value.
			// On a black background with a white watermark, max(R,G,B) gives us
			// the effective brightness, which equals the alpha value.
			maxChannel := r8
			if g8 > maxChannel {
				maxChannel = g8
			}
			if b8 > maxChannel {
				maxChannel = b8
			}

			// Normalize to [0.0, 1.0] range and store
			alphaMap[y*width+x] = float32(maxChannel) / 255.0
		}
	}

	return alphaMap
}
