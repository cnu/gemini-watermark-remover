package watermark

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

// Algorithm constants for watermark removal.
// These values are tuned to match Gemini's watermark application.
const (
	// AlphaThreshold is the minimum alpha value to process.
	// Pixels with alpha below this are considered transparent and skipped.
	// This avoids unnecessary processing and potential numerical issues.
	AlphaThreshold = 0.002

	// MaxAlpha is the maximum alpha value allowed during processing.
	// Values above this are clamped to prevent division by near-zero
	// in the reverse blending formula (1 - alpha would be too small).
	MaxAlpha = 0.99

	// LogoValue is the color value of the Gemini watermark logo.
	// The logo is white, so all RGB channels are 255.
	LogoValue = 255.0
)

// WatermarkConfig holds the size and margin configuration for a watermark.
// Gemini uses different configurations based on image dimensions.
type WatermarkConfig struct {
	// Size is the width and height of the watermark in pixels (48 or 96).
	Size int

	// Margin is the distance from the image edge in pixels (32 or 64).
	Margin int
}

// Engine handles watermark removal with pre-computed alpha maps.
// Create an Engine once and reuse it for multiple images to avoid
// recalculating alpha maps.
type Engine struct {
	// alphaMap48 contains alpha values for the 48x48 watermark.
	// Each value is in the range [0.0, 1.0].
	alphaMap48 []float32

	// alphaMap96 contains alpha values for the 96x96 watermark.
	// Each value is in the range [0.0, 1.0].
	alphaMap96 []float32
}

// NewEngine creates a new watermark removal engine.
// It loads the embedded reference images and pre-computes the alpha maps.
// Returns an error if the reference images cannot be loaded.
func NewEngine() (*Engine, error) {
	// Load and calculate alpha map for 48x48 watermark (used for smaller images)
	img48, err := LoadReferenceImage(48)
	if err != nil {
		return nil, fmt.Errorf("failed to load 48px reference: %w", err)
	}
	alphaMap48 := CalculateAlphaMap(img48)

	// Load and calculate alpha map for 96x96 watermark (used for larger images)
	img96, err := LoadReferenceImage(96)
	if err != nil {
		return nil, fmt.Errorf("failed to load 96px reference: %w", err)
	}
	alphaMap96 := CalculateAlphaMap(img96)

	return &Engine{
		alphaMap48: alphaMap48,
		alphaMap96: alphaMap96,
	}, nil
}

// DetectConfig determines the watermark configuration based on image dimensions.
// Gemini uses a larger watermark (96x96) for images where both dimensions
// exceed 1024 pixels, and a smaller one (48x48) for everything else.
func DetectConfig(width, height int) WatermarkConfig {
	if width > 1024 && height > 1024 {
		return WatermarkConfig{Size: 96, Margin: 64}
	}
	return WatermarkConfig{Size: 48, Margin: 32}
}

// CalculatePosition returns the rectangle where the watermark is located.
// The watermark is always positioned in the bottom-right corner of the image,
// offset by the margin value from each edge.
func CalculatePosition(imgWidth, imgHeight int, config WatermarkConfig) image.Rectangle {
	// Calculate top-left corner of watermark region
	x := imgWidth - config.Margin - config.Size
	y := imgHeight - config.Margin - config.Size
	return image.Rect(x, y, x+config.Size, y+config.Size)
}

// RemoveWatermark removes the Gemini watermark from an image.
// It automatically detects the appropriate watermark size based on
// the image dimensions and applies reverse alpha blending to restore
// the original pixels.
//
// The function returns a new image with the watermark removed.
// The original image is not modified.
func (e *Engine) RemoveWatermark(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a new RGBA image and copy the source into it.
	// We work on a copy to avoid modifying the original.
	result := image.NewRGBA(bounds)
	draw.Draw(result, bounds, img, bounds.Min, draw.Src)

	// Detect watermark configuration based on image size
	config := DetectConfig(width, height)
	position := CalculatePosition(width, height, config)

	// Select the appropriate pre-computed alpha map
	var alphaMap []float32
	if config.Size == 96 {
		alphaMap = e.alphaMap96
	} else {
		alphaMap = e.alphaMap48
	}

	// Process each pixel in the watermark region.
	// Apply the reverse alpha blending formula to recover original colors.
	for row := 0; row < config.Size; row++ {
		for col := 0; col < config.Size; col++ {
			// Calculate image coordinates for this watermark pixel
			imgX := position.Min.X + col
			imgY := position.Min.Y + row

			// Skip pixels outside image bounds (edge cases)
			if imgX < 0 || imgY < 0 || imgX >= width || imgY >= height {
				continue
			}

			// Get alpha value from pre-computed map
			alphaIdx := row*config.Size + col
			alpha := alphaMap[alphaIdx]

			// Skip nearly-transparent pixels (no watermark effect here)
			if alpha < AlphaThreshold {
				continue
			}

			// Clamp alpha to prevent division by values too close to zero.
			// When alpha approaches 1.0, (1 - alpha) approaches 0, causing
			// numerical instability in the division.
			if alpha > MaxAlpha {
				alpha = MaxAlpha
			}

			// Pre-compute (1 - alpha) for the reverse blending formula
			oneMinusAlpha := 1.0 - float64(alpha)

			// Get the current (watermarked) pixel values.
			// RGBA() returns values in [0, 65535], so we shift to get [0, 255].
			r, g, b, a := result.At(imgX, imgY).RGBA()
			watermarkedR := float64(r >> 8)
			watermarkedG := float64(g >> 8)
			watermarkedB := float64(b >> 8)

			// Apply reverse alpha blending formula:
			// original = (watermarked - alpha * logo) / (1 - alpha)
			//
			// This inverts the formula Gemini used to apply the watermark:
			// watermarked = alpha * logo + (1 - alpha) * original
			alphaF := float64(alpha)
			originalR := (watermarkedR - alphaF*LogoValue) / oneMinusAlpha
			originalG := (watermarkedG - alphaF*LogoValue) / oneMinusAlpha
			originalB := (watermarkedB - alphaF*LogoValue) / oneMinusAlpha

			// Clamp results to valid 8-bit range [0, 255].
			// Values can go out of range due to JPEG compression artifacts
			// or slight variations in the watermark application.
			originalR = clamp(originalR, 0, 255)
			originalG = clamp(originalG, 0, 255)
			originalB = clamp(originalB, 0, 255)

			// Write the restored pixel back to the result image
			result.SetRGBA(imgX, imgY, color.RGBA{
				R: uint8(originalR),
				G: uint8(originalG),
				B: uint8(originalB),
				A: uint8(a >> 8), // Preserve original alpha
			})
		}
	}

	return result
}

// GetWatermarkInfo returns information about the watermark configuration
// and position for a given image size. Useful for debugging or displaying
// information to the user.
func GetWatermarkInfo(width, height int) (config WatermarkConfig, position image.Rectangle) {
	config = DetectConfig(width, height)
	position = CalculatePosition(width, height, config)
	return
}

// clamp restricts a value to the range [min, max].
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
