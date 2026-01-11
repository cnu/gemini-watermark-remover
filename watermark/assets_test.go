package watermark

import (
	"testing"
)

func TestLoadReferenceImage_48(t *testing.T) {
	img, err := LoadReferenceImage(48)
	if err != nil {
		t.Fatalf("LoadReferenceImage(48) error: %v", err)
	}
	if img == nil {
		t.Fatal("LoadReferenceImage(48) returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 48 || bounds.Dy() != 48 {
		t.Errorf("expected 48x48 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestLoadReferenceImage_96(t *testing.T) {
	img, err := LoadReferenceImage(96)
	if err != nil {
		t.Fatalf("LoadReferenceImage(96) error: %v", err)
	}
	if img == nil {
		t.Fatal("LoadReferenceImage(96) returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 96 || bounds.Dy() != 96 {
		t.Errorf("expected 96x96 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestLoadReferenceImage_DefaultsTo48(t *testing.T) {
	// Any size other than 48 or 96 should default to 48
	testSizes := []int{0, 1, 32, 64, 100, -1}

	for _, size := range testSizes {
		img, err := LoadReferenceImage(size)
		if err != nil {
			t.Fatalf("LoadReferenceImage(%d) error: %v", size, err)
		}

		bounds := img.Bounds()
		if bounds.Dx() != 48 || bounds.Dy() != 48 {
			t.Errorf("LoadReferenceImage(%d): expected 48x48 (default), got %dx%d",
				size, bounds.Dx(), bounds.Dy())
		}
	}
}

func TestEmbeddedAssetsNotEmpty(t *testing.T) {
	// Verify the embedded assets have content
	if len(bg48PNG) == 0 {
		t.Error("bg48PNG is empty")
	}
	if len(bg96PNG) == 0 {
		t.Error("bg96PNG is empty")
	}

	// Verify they have PNG signature (first 8 bytes)
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	if len(bg48PNG) < 8 {
		t.Error("bg48PNG too small to be valid PNG")
	} else {
		for i, b := range pngSignature {
			if bg48PNG[i] != b {
				t.Errorf("bg48PNG: invalid PNG signature at byte %d", i)
				break
			}
		}
	}

	if len(bg96PNG) < 8 {
		t.Error("bg96PNG too small to be valid PNG")
	} else {
		for i, b := range pngSignature {
			if bg96PNG[i] != b {
				t.Errorf("bg96PNG: invalid PNG signature at byte %d", i)
				break
			}
		}
	}
}

func TestReferenceImageHasContent(t *testing.T) {
	// The reference images should have non-zero pixels (the watermark)
	// and zero pixels (the black background)

	testCases := []int{48, 96}

	for _, size := range testCases {
		img, err := LoadReferenceImage(size)
		if err != nil {
			t.Fatalf("LoadReferenceImage(%d) error: %v", size, err)
		}

		bounds := img.Bounds()
		hasNonZero := false
		hasZero := false

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				if r == 0 && g == 0 && b == 0 {
					hasZero = true
				} else {
					hasNonZero = true
				}

				if hasZero && hasNonZero {
					break
				}
			}
			if hasZero && hasNonZero {
				break
			}
		}

		if !hasNonZero {
			t.Errorf("reference image %d: no non-zero pixels (no watermark content)", size)
		}
		if !hasZero {
			t.Errorf("reference image %d: no zero pixels (no black background)", size)
		}
	}
}
