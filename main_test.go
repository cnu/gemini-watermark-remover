package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsGlobPattern(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		// Patterns with glob characters
		{"*.png", true},
		{"*.jpg", true},
		{"images/*.png", true},
		{"photos/**/*.jpg", true},
		{"file?.png", true},
		{"image[0-9].png", true},
		{"[abc].png", true},

		// Not glob patterns
		{"image.png", false},
		{"path/to/image.png", false},
		{"my-file_name.jpg", false},
		{"./images/", false},
		{"/absolute/path/file.png", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := isGlobPattern(tc.input)
		if result != tc.expected {
			t.Errorf("isGlobPattern(%q) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}

func TestIsSupportedImage(t *testing.T) {
	testCases := []struct {
		filename string
		expected bool
	}{
		// Supported formats
		{"image.png", true},
		{"image.PNG", true},
		{"image.Png", true},
		{"photo.jpg", true},
		{"photo.JPG", true},
		{"photo.jpeg", true},
		{"photo.JPEG", true},
		{"path/to/image.png", true},
		{"my-file_name.jpg", true},

		// Unsupported formats
		{"image.gif", false},
		{"image.webp", false},
		{"image.bmp", false},
		{"image.tiff", false},
		{"image.svg", false},
		{"document.pdf", false},
		{"file.txt", false},
		{"noextension", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := isSupportedImage(tc.filename)
		if result != tc.expected {
			t.Errorf("isSupportedImage(%q) = %v, expected %v", tc.filename, result, tc.expected)
		}
	}
}

func TestGenerateOutputPath(t *testing.T) {
	testCases := []struct {
		inputPath string
		suffix    string
		expected  string
	}{
		{"image.png", "_clean", "image_clean.png"},
		{"image.jpg", "_clean", "image_clean.jpg"},
		{"photo.jpeg", "_processed", "photo_processed.jpeg"},
		{"path/to/image.png", "_clean", "path/to/image_clean.png"},
		{"./image.png", "_clean", "image_clean.png"},
		{"my-file_name.png", "_out", "my-file_name_out.png"},
	}

	for _, tc := range testCases {
		result := generateOutputPath(tc.inputPath, tc.suffix)
		if result != tc.expected {
			t.Errorf("generateOutputPath(%q, %q) = %q, expected %q",
				tc.inputPath, tc.suffix, result, tc.expected)
		}
	}
}

func TestExpandGlob(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir, err := os.MkdirTemp("", "glob_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := []string{
		"image1.png",
		"image2.png",
		"photo.jpg",
		"document.txt",
		"image_clean.png", // Should be skipped (has suffix)
	}

	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", f, err)
		}
	}

	// Save original suffix and set test suffix
	originalSuffix := suffix
	suffix = "_clean"
	defer func() { suffix = originalSuffix }()

	// Test glob expansion for PNG files
	pattern := filepath.Join(tmpDir, "*.png")
	files, err := expandGlob(pattern)
	if err != nil {
		t.Fatalf("expandGlob(%q) error: %v", pattern, err)
	}

	// Should find image1.png and image2.png, but not image_clean.png
	if len(files) != 2 {
		t.Errorf("expandGlob(%q) returned %d files, expected 2", pattern, len(files))
	}

	// Test glob expansion for all images
	pattern = filepath.Join(tmpDir, "*.*")
	files, err = expandGlob(pattern)
	if err != nil {
		t.Fatalf("expandGlob(%q) error: %v", pattern, err)
	}

	// Should find image1.png, image2.png, photo.jpg (3 files)
	// Excludes document.txt (not an image) and image_clean.png (has suffix)
	if len(files) != 3 {
		t.Errorf("expandGlob(%q) returned %d files, expected 3", pattern, len(files))
	}

	// Test non-matching pattern
	pattern = filepath.Join(tmpDir, "*.gif")
	files, err = expandGlob(pattern)
	if err != nil {
		t.Fatalf("expandGlob(%q) error: %v", pattern, err)
	}

	if len(files) != 0 {
		t.Errorf("expandGlob(%q) returned %d files, expected 0", pattern, len(files))
	}
}

func TestFindImageFiles(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir, err := os.MkdirTemp("", "find_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := []string{
		"image1.png",
		"image2.PNG", // uppercase extension
		"photo.jpg",
		"photo2.JPEG",
		"document.txt",
		"image_clean.png", // Should be skipped (has suffix)
	}

	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", f, err)
		}
	}

	// Create a subdirectory (should be skipped)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Save original suffix and set test suffix
	originalSuffix := suffix
	suffix = "_clean"
	defer func() { suffix = originalSuffix }()

	files, err := findImageFiles(tmpDir)
	if err != nil {
		t.Fatalf("findImageFiles(%q) error: %v", tmpDir, err)
	}

	// Should find: image1.png, image2.PNG, photo.jpg, photo2.JPEG (4 files)
	// Excludes: document.txt (not image), image_clean.png (has suffix), subdir (directory)
	if len(files) != 4 {
		t.Errorf("findImageFiles(%q) returned %d files, expected 4", tmpDir, len(files))
		for _, f := range files {
			t.Logf("  Found: %s", f)
		}
	}
}

func TestFindImageFiles_EmptyDirectory(t *testing.T) {
	// Create an empty temporary directory
	tmpDir, err := os.MkdirTemp("", "empty_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	files, err := findImageFiles(tmpDir)
	if err != nil {
		t.Fatalf("findImageFiles(%q) error: %v", tmpDir, err)
	}

	if len(files) != 0 {
		t.Errorf("findImageFiles on empty dir returned %d files, expected 0", len(files))
	}
}

func TestFindImageFiles_NonexistentDirectory(t *testing.T) {
	_, err := findImageFiles("/nonexistent/directory/path")
	if err == nil {
		t.Error("findImageFiles on nonexistent directory should return error")
	}
}

func TestSuffixDetectionInFilename(t *testing.T) {
	// Test that files with suffix are correctly detected
	testCases := []struct {
		filename string
		suffix   string
		expected bool // true if should be skipped
	}{
		{"image_clean.png", "_clean", true},
		{"image.png", "_clean", false},
		{"photo_CLEAN.jpg", "_clean", true}, // case insensitive
		{"image_clean_v2.png", "_clean", true},
		{"cleanimage.png", "_clean", false}, // suffix not as suffix
		{"my_clean_file.png", "_clean", true},
	}

	for _, tc := range testCases {
		contains := strings.Contains(strings.ToLower(tc.filename), strings.ToLower(tc.suffix))
		if contains != tc.expected {
			t.Errorf("filename %q with suffix %q: contains=%v, expected=%v",
				tc.filename, tc.suffix, contains, tc.expected)
		}
	}
}

func TestExpandGlob_SkipsDirectories(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "glob_dir_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a subdirectory that matches the glob pattern
	subDir := filepath.Join(tmpDir, "images.png") // directory with .png in name
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create an actual file
	filePath := filepath.Join(tmpDir, "photo.png")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Save original suffix and set test suffix
	originalSuffix := suffix
	suffix = "_clean"
	defer func() { suffix = originalSuffix }()

	pattern := filepath.Join(tmpDir, "*.png")
	files, err := expandGlob(pattern)
	if err != nil {
		t.Fatalf("expandGlob error: %v", err)
	}

	// Should only find the file, not the directory
	if len(files) != 1 {
		t.Errorf("expandGlob returned %d files, expected 1 (should skip directory)", len(files))
	}
}
