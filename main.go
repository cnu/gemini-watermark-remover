// Gemini Watermark Remover
//
// A command-line tool to remove Gemini AI watermarks from generated images.
// Supports single files, multiple files, glob patterns, or batch processing of directories.
//
// Usage:
//
//	gemini-watermark-remover [options] <files|directories|globs>...
//
// Examples:
//
//	gemini-watermark-remover image.png                    # Process single image
//	gemini-watermark-remover img1.png img2.png img3.jpg   # Process multiple files
//	gemini-watermark-remover ./images/                    # Process all images in folder
//	gemini-watermark-remover "*.png"                      # Process all PNG files (glob)
//	gemini-watermark-remover "photos/*.jpg"               # Glob with directory
//	gemini-watermark-remover -v -s _clean image.png       # Verbose with custom suffix
package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"gemini-watermark-remover/watermark"
)

// Command-line flags
var (
	// suffix is appended to the output filename (before the extension)
	suffix string

	// verbose enables detailed output about watermark detection and processing
	verbose bool

	// quiet suppresses all output except errors
	quiet bool
)

func main() {
	// Define command-line flags with both short and long versions
	flag.StringVar(&suffix, "s", "_clean", "Suffix to append to output filename")
	flag.StringVar(&suffix, "suffix", "_clean", "Suffix to append to output filename")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&quiet, "q", false, "Suppress all output except errors")
	flag.BoolVar(&quiet, "quiet", false, "Suppress all output except errors")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Gemini Watermark Remover\n\n")
		fmt.Fprintf(os.Stderr, "Removes the Gemini AI watermark from generated images using\n")
		fmt.Fprintf(os.Stderr, "reverse alpha blending.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <files|directories|globs>...\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s image.png                    # Process single image\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s img1.png img2.jpg img3.png   # Process multiple files\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -s _nowm image.png           # Custom suffix\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s ./images/                    # Process all images in folder\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s \"*.png\"                      # Process all PNG files (glob)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s \"photos/*.jpg\" ./other/      # Mix glob and directory\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -v ./images/                 # Verbose mode\n", os.Args[0])
	}

	flag.Parse()

	// Require at least one positional argument (file, directory, or glob pattern)
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Initialize the watermark removal engine.
	// This loads and pre-processes the reference watermark images.
	engine, err := watermark.NewEngine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing engine: %v\n", err)
		os.Exit(1)
	}

	// Build list of files to process from all arguments
	var files []string

	for _, inputPath := range flag.Args() {
		// Check if input looks like a glob pattern
		if isGlobPattern(inputPath) {
			matched, err := expandGlob(inputPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error expanding glob pattern %s: %v\n", inputPath, err)
				continue
			}
			files = append(files, matched...)
		} else {
			// Check if it's a file or directory
			info, err := os.Stat(inputPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error accessing path %s: %v\n", inputPath, err)
				continue
			}

			if info.IsDir() {
				// Scan directory for image files
				dirFiles, err := findImageFiles(inputPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error scanning directory %s: %v\n", inputPath, err)
					continue
				}
				files = append(files, dirFiles...)
			} else {
				// Single file - skip if it already has the output suffix
				if strings.Contains(strings.ToLower(inputPath), strings.ToLower(suffix)) {
					if !quiet {
						fmt.Printf("Skipping %s (already processed)\n", inputPath)
					}
					continue
				}
				files = append(files, inputPath)
			}
		}
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No image files found to process\n")
		os.Exit(1)
	}

	if !quiet {
		fmt.Printf("Found %d image(s) to process\n", len(files))
	}

	// Process each file and track success count
	successCount := 0
	for _, file := range files {
		err := processImage(engine, file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file, err)
			continue
		}
		successCount++
	}

	// Print summary
	if !quiet {
		fmt.Printf("Successfully processed %d/%d image(s)\n", successCount, len(files))
	}
}

// isGlobPattern checks if the input string contains glob metacharacters.
func isGlobPattern(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}

// expandGlob expands a glob pattern and returns matching image files.
// It filters results to only include supported image formats (PNG, JPEG)
// and excludes files that already have the output suffix.
func expandGlob(pattern string) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, match := range matches {
		// Skip directories
		info, err := os.Stat(match)
		if err != nil || info.IsDir() {
			continue
		}

		// Check if it's a supported image format
		if !isSupportedImage(match) {
			continue
		}

		// Skip files that already have our output suffix
		if strings.Contains(strings.ToLower(match), strings.ToLower(suffix)) {
			continue
		}

		files = append(files, match)
	}

	return files, nil
}

// isSupportedImage checks if a file has a supported image extension.
func isSupportedImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}

// findImageFiles scans a directory for supported image files (PNG, JPEG).
// It returns a list of absolute file paths. Files that already contain
// the suffix in their name are skipped to avoid reprocessing.
//
// Note: This function is non-recursive and only scans the immediate directory.
func findImageFiles(dir string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Check if it's a supported image format
		if !isSupportedImage(name) {
			continue
		}

		// Skip files that already have our output suffix to avoid
		// reprocessing previously cleaned images
		if strings.Contains(strings.ToLower(name), strings.ToLower(suffix)) {
			continue
		}

		files = append(files, filepath.Join(dir, name))
	}

	return files, nil
}

// processImage loads an image, removes the watermark, and saves the result.
// The output filename is the same as input with the suffix appended before
// the extension (e.g., "photo.png" -> "photo_clean.png").
//
// The output format matches the input format:
//   - PNG input produces PNG output (lossless)
//   - JPEG input produces JPEG output (95% quality)
func processImage(engine *watermark.Engine, inputPath string) error {
	// Open the input file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode the image. The format is automatically detected from the header.
	// Supported formats: PNG, JPEG (registered via image/png and image/jpeg imports)
	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// In verbose mode, display watermark detection information
	if verbose {
		bounds := img.Bounds()
		config, pos := watermark.GetWatermarkInfo(bounds.Dx(), bounds.Dy())
		fmt.Printf("Processing: %s (%dx%d, format: %s)\n", inputPath, bounds.Dx(), bounds.Dy(), format)
		fmt.Printf("  Watermark: %dx%d at position (%d, %d)\n", config.Size, config.Size, pos.Min.X, pos.Min.Y)
	}

	// Remove the watermark using reverse alpha blending
	result := engine.RemoveWatermark(img)

	// Generate output path with suffix
	outputPath := generateOutputPath(inputPath, suffix)

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Encode in the same format as input to preserve quality characteristics.
	// PNG remains lossless, JPEG uses high quality (95%).
	switch format {
	case "png":
		err = png.Encode(outFile, result)
	case "jpeg":
		err = jpeg.Encode(outFile, result, &jpeg.Options{Quality: 95})
	default:
		// Unknown format - default to PNG for safety (lossless)
		err = png.Encode(outFile, result)
	}

	if err != nil {
		return fmt.Errorf("failed to encode output: %w", err)
	}

	if !quiet {
		fmt.Printf("Saved: %s\n", outputPath)
	}

	return nil
}

// generateOutputPath creates the output filename by inserting a suffix
// before the file extension.
//
// Example: generateOutputPath("/path/to/image.png", "_clean") returns
// "/path/to/image_clean.png"
func generateOutputPath(inputPath, suffix string) string {
	dir := filepath.Dir(inputPath)
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(filepath.Base(inputPath), ext)
	return filepath.Join(dir, base+suffix+ext)
}
