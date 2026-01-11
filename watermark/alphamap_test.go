package watermark

import (
	"image"
	"image/color"
	"testing"
)

// createTestImage creates a simple test image with the given dimensions and color.
func createTestImage(width, height int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestCalculateAlphaMap_BlackImage(t *testing.T) {
	// A pure black image should produce all zero alpha values
	img := createTestImage(4, 4, color.Black)
	alphaMap := CalculateAlphaMap(img)

	if len(alphaMap) != 16 {
		t.Errorf("expected alpha map length 16, got %d", len(alphaMap))
	}

	for i, alpha := range alphaMap {
		if alpha != 0.0 {
			t.Errorf("pixel %d: expected alpha 0.0, got %f", i, alpha)
		}
	}
}

func TestCalculateAlphaMap_WhiteImage(t *testing.T) {
	// A pure white image should produce all 1.0 alpha values
	img := createTestImage(4, 4, color.White)
	alphaMap := CalculateAlphaMap(img)

	if len(alphaMap) != 16 {
		t.Errorf("expected alpha map length 16, got %d", len(alphaMap))
	}

	for i, alpha := range alphaMap {
		if alpha != 1.0 {
			t.Errorf("pixel %d: expected alpha 1.0, got %f", i, alpha)
		}
	}
}

func TestCalculateAlphaMap_GrayImage(t *testing.T) {
	// A 50% gray image should produce ~0.5 alpha values
	gray := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	img := createTestImage(4, 4, gray)
	alphaMap := CalculateAlphaMap(img)

	expectedAlpha := float32(128) / 255.0

	for i, alpha := range alphaMap {
		if alpha != expectedAlpha {
			t.Errorf("pixel %d: expected alpha %f, got %f", i, expectedAlpha, alpha)
		}
	}
}

func TestCalculateAlphaMap_MaxChannel(t *testing.T) {
	// Test that max(R,G,B) is used, not average
	// Red channel is highest
	redDominant := color.RGBA{R: 200, G: 100, B: 50, A: 255}
	img := createTestImage(2, 2, redDominant)
	alphaMap := CalculateAlphaMap(img)

	expectedAlpha := float32(200) / 255.0

	for i, alpha := range alphaMap {
		if alpha != expectedAlpha {
			t.Errorf("pixel %d: expected alpha %f (from max channel), got %f", i, expectedAlpha, alpha)
		}
	}
}

func TestCalculateAlphaMap_Dimensions(t *testing.T) {
	// Test various image dimensions
	testCases := []struct {
		width, height int
	}{
		{1, 1},
		{10, 10},
		{48, 48},
		{96, 96},
		{100, 50},
	}

	for _, tc := range testCases {
		img := createTestImage(tc.width, tc.height, color.Gray{Y: 128})
		alphaMap := CalculateAlphaMap(img)

		expectedLen := tc.width * tc.height
		if len(alphaMap) != expectedLen {
			t.Errorf("dimensions %dx%d: expected length %d, got %d",
				tc.width, tc.height, expectedLen, len(alphaMap))
		}
	}
}

func TestCalculateAlphaMap_RowMajorOrder(t *testing.T) {
	// Create an image with a gradient to verify row-major ordering
	img := image.NewRGBA(image.Rect(0, 0, 3, 2))

	// Row 0: values 0, 85, 170
	img.Set(0, 0, color.Gray{Y: 0})
	img.Set(1, 0, color.Gray{Y: 85})
	img.Set(2, 0, color.Gray{Y: 170})

	// Row 1: values 255, 128, 64
	img.Set(0, 1, color.Gray{Y: 255})
	img.Set(1, 1, color.Gray{Y: 128})
	img.Set(2, 1, color.Gray{Y: 64})

	alphaMap := CalculateAlphaMap(img)

	expected := []float32{
		0.0 / 255.0, 85.0 / 255.0, 170.0 / 255.0, // Row 0
		255.0 / 255.0, 128.0 / 255.0, 64.0 / 255.0, // Row 1
	}

	for i, exp := range expected {
		if alphaMap[i] != exp {
			t.Errorf("index %d: expected %f, got %f", i, exp, alphaMap[i])
		}
	}
}
