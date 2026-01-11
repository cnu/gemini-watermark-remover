package watermark

import (
	"image"
	"image/color"
	"testing"
)

func TestDetectConfig_LargeImage(t *testing.T) {
	// Images larger than 1024x1024 should use 96px watermark with 64px margin
	testCases := []struct {
		width, height int
	}{
		{1025, 1025},
		{2000, 2000},
		{1920, 1080}, // This is NOT > 1024 in both dimensions
		{1080, 1920}, // This is NOT > 1024 in both dimensions
		{4096, 4096},
	}

	for _, tc := range testCases {
		config := DetectConfig(tc.width, tc.height)

		// Both dimensions must exceed 1024 for large watermark
		expectLarge := tc.width > 1024 && tc.height > 1024

		if expectLarge {
			if config.Size != 96 {
				t.Errorf("dimensions %dx%d: expected size 96, got %d", tc.width, tc.height, config.Size)
			}
			if config.Margin != 64 {
				t.Errorf("dimensions %dx%d: expected margin 64, got %d", tc.width, tc.height, config.Margin)
			}
		} else {
			if config.Size != 48 {
				t.Errorf("dimensions %dx%d: expected size 48, got %d", tc.width, tc.height, config.Size)
			}
			if config.Margin != 32 {
				t.Errorf("dimensions %dx%d: expected margin 32, got %d", tc.width, tc.height, config.Margin)
			}
		}
	}
}

func TestDetectConfig_SmallImage(t *testing.T) {
	// Images 1024x1024 or smaller should use 48px watermark with 32px margin
	testCases := []struct {
		width, height int
	}{
		{100, 100},
		{512, 512},
		{1024, 1024},
		{800, 600},
		{1024, 2000}, // Only one dimension > 1024
		{2000, 1024}, // Only one dimension > 1024
	}

	for _, tc := range testCases {
		config := DetectConfig(tc.width, tc.height)

		if config.Size != 48 {
			t.Errorf("dimensions %dx%d: expected size 48, got %d", tc.width, tc.height, config.Size)
		}
		if config.Margin != 32 {
			t.Errorf("dimensions %dx%d: expected margin 32, got %d", tc.width, tc.height, config.Margin)
		}
	}
}

func TestCalculatePosition_SmallConfig(t *testing.T) {
	config := WatermarkConfig{Size: 48, Margin: 32}

	testCases := []struct {
		imgWidth, imgHeight int
		expectedX, expectedY int
	}{
		{800, 600, 800 - 32 - 48, 600 - 32 - 48},   // 720, 520
		{1024, 768, 1024 - 32 - 48, 768 - 32 - 48}, // 944, 688
		{100, 100, 100 - 32 - 48, 100 - 32 - 48},   // 20, 20
	}

	for _, tc := range testCases {
		pos := CalculatePosition(tc.imgWidth, tc.imgHeight, config)

		if pos.Min.X != tc.expectedX {
			t.Errorf("image %dx%d: expected X=%d, got %d",
				tc.imgWidth, tc.imgHeight, tc.expectedX, pos.Min.X)
		}
		if pos.Min.Y != tc.expectedY {
			t.Errorf("image %dx%d: expected Y=%d, got %d",
				tc.imgWidth, tc.imgHeight, tc.expectedY, pos.Min.Y)
		}
		if pos.Dx() != config.Size {
			t.Errorf("image %dx%d: expected width %d, got %d",
				tc.imgWidth, tc.imgHeight, config.Size, pos.Dx())
		}
		if pos.Dy() != config.Size {
			t.Errorf("image %dx%d: expected height %d, got %d",
				tc.imgWidth, tc.imgHeight, config.Size, pos.Dy())
		}
	}
}

func TestCalculatePosition_LargeConfig(t *testing.T) {
	config := WatermarkConfig{Size: 96, Margin: 64}

	testCases := []struct {
		imgWidth, imgHeight int
		expectedX, expectedY int
	}{
		{2000, 2000, 2000 - 64 - 96, 2000 - 64 - 96}, // 1840, 1840
		{1920, 1080, 1920 - 64 - 96, 1080 - 64 - 96}, // 1760, 920
	}

	for _, tc := range testCases {
		pos := CalculatePosition(tc.imgWidth, tc.imgHeight, config)

		if pos.Min.X != tc.expectedX {
			t.Errorf("image %dx%d: expected X=%d, got %d",
				tc.imgWidth, tc.imgHeight, tc.expectedX, pos.Min.X)
		}
		if pos.Min.Y != tc.expectedY {
			t.Errorf("image %dx%d: expected Y=%d, got %d",
				tc.imgWidth, tc.imgHeight, tc.expectedY, pos.Min.Y)
		}
	}
}

func TestGetWatermarkInfo(t *testing.T) {
	// Test that GetWatermarkInfo returns consistent results with DetectConfig and CalculatePosition
	testCases := []struct {
		width, height int
	}{
		{800, 600},
		{1024, 768},
		{2000, 2000},
	}

	for _, tc := range testCases {
		config, pos := GetWatermarkInfo(tc.width, tc.height)

		expectedConfig := DetectConfig(tc.width, tc.height)
		expectedPos := CalculatePosition(tc.width, tc.height, expectedConfig)

		if config != expectedConfig {
			t.Errorf("dimensions %dx%d: config mismatch", tc.width, tc.height)
		}
		if pos != expectedPos {
			t.Errorf("dimensions %dx%d: position mismatch", tc.width, tc.height)
		}
	}
}

func TestClamp(t *testing.T) {
	testCases := []struct {
		value, min, max float64
		expected        float64
	}{
		{50, 0, 100, 50},    // Within range
		{-10, 0, 100, 0},    // Below min
		{150, 0, 100, 100},  // Above max
		{0, 0, 100, 0},      // At min
		{100, 0, 100, 100},  // At max
		{0.5, 0, 1, 0.5},    // Floating point
		{-0.5, 0, 1, 0},     // Floating point below
		{1.5, 0, 1, 1},      // Floating point above
	}

	for _, tc := range testCases {
		result := clamp(tc.value, tc.min, tc.max)
		if result != tc.expected {
			t.Errorf("clamp(%f, %f, %f) = %f, expected %f",
				tc.value, tc.min, tc.max, result, tc.expected)
		}
	}
}

func TestNewEngine(t *testing.T) {
	// Test that engine can be created successfully
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() returned error: %v", err)
	}
	if engine == nil {
		t.Fatal("NewEngine() returned nil engine")
	}

	// Verify alpha maps are populated
	if len(engine.alphaMap48) != 48*48 {
		t.Errorf("expected alphaMap48 length %d, got %d", 48*48, len(engine.alphaMap48))
	}
	if len(engine.alphaMap96) != 96*96 {
		t.Errorf("expected alphaMap96 length %d, got %d", 96*96, len(engine.alphaMap96))
	}
}

func TestRemoveWatermark_PreservesImageDimensions(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}

	testCases := []struct {
		width, height int
	}{
		{800, 600},
		{1024, 768},
		{2000, 1500},
	}

	for _, tc := range testCases {
		img := image.NewRGBA(image.Rect(0, 0, tc.width, tc.height))
		// Fill with a solid color
		for y := 0; y < tc.height; y++ {
			for x := 0; x < tc.width; x++ {
				img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
			}
		}

		result := engine.RemoveWatermark(img)

		if result.Bounds().Dx() != tc.width {
			t.Errorf("dimensions %dx%d: result width %d != input width %d",
				tc.width, tc.height, result.Bounds().Dx(), tc.width)
		}
		if result.Bounds().Dy() != tc.height {
			t.Errorf("dimensions %dx%d: result height %d != input height %d",
				tc.width, tc.height, result.Bounds().Dy(), tc.height)
		}
	}
}

func TestRemoveWatermark_PreservesNonWatermarkRegion(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}

	// Create a small image where we can verify pixels outside watermark region
	width, height := 200, 200
	testColor := color.RGBA{R: 100, G: 150, B: 200, A: 255}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, testColor)
		}
	}

	result := engine.RemoveWatermark(img)

	// Check a pixel far from the watermark region (top-left corner)
	r, g, b, a := result.At(0, 0).RGBA()
	r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)

	if r8 != testColor.R || g8 != testColor.G || b8 != testColor.B || a8 != testColor.A {
		t.Errorf("non-watermark pixel changed: got (%d,%d,%d,%d), expected (%d,%d,%d,%d)",
			r8, g8, b8, a8, testColor.R, testColor.G, testColor.B, testColor.A)
	}
}

func TestConstants(t *testing.T) {
	// Verify constants are within expected ranges
	if AlphaThreshold <= 0 || AlphaThreshold >= 1 {
		t.Errorf("AlphaThreshold should be between 0 and 1, got %f", AlphaThreshold)
	}

	if MaxAlpha <= 0 || MaxAlpha >= 1 {
		t.Errorf("MaxAlpha should be between 0 and 1, got %f", MaxAlpha)
	}

	if MaxAlpha <= AlphaThreshold {
		t.Errorf("MaxAlpha (%f) should be greater than AlphaThreshold (%f)",
			MaxAlpha, AlphaThreshold)
	}

	if LogoValue != 255.0 {
		t.Errorf("LogoValue should be 255.0 (white), got %f", LogoValue)
	}
}
