# Gemini Watermark Remover

A command-line tool written in Go that removes the Gemini AI watermark from generated images using reverse alpha blending.

## How It Works

Google's Gemini AI adds a semi-transparent watermark to generated images in the bottom-right corner. This tool removes it by reversing the alpha compositing formula.

### The Math

When Gemini applies the watermark, it uses standard alpha blending:

```
watermarked_pixel = alpha * logo_color + (1 - alpha) * original_pixel
```

Since the logo is white (255) and we know the alpha values from reference images, we can solve for the original:

```
original_pixel = (watermarked_pixel - alpha * 255) / (1 - alpha)
```

### Watermark Specifications

| Image Size | Watermark Size | Margin from Edge |
|------------|----------------|------------------|
| > 1024x1024 | 96x96 pixels | 64 pixels |
| <= 1024x1024 | 48x48 pixels | 32 pixels |

## Installation

### From Source

Requires Go 1.16+ (for embed support).

```bash
git clone https://github.com/yourusername/gemini-watermark-remover.git
cd gemini-watermark-remover
go build -o gemini-watermark-remover .
```

### Pre-built Binary

Download the latest release from the releases page.

## Usage

```bash
# Process a single image
./gemini-watermark-remover image.png

# Process all images in a directory
./gemini-watermark-remover ./my-images/

# Use a custom suffix (default is "_clean")
./gemini-watermark-remover -s "_nowatermark" image.png

# Verbose mode - shows watermark detection info
./gemini-watermark-remover -v image.png

# Quiet mode - only show errors
./gemini-watermark-remover -q ./my-images/
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-s`, `--suffix` | Suffix added to output filename | `_clean` |
| `-v`, `--verbose` | Show detailed processing information | `false` |
| `-q`, `--quiet` | Suppress all output except errors | `false` |

### Output

- Output files are saved in the same directory as the input
- Original format is preserved (PNG -> PNG, JPEG -> JPEG)
- JPEG output uses 95% quality

### Examples

```bash
# Input: photo.png
# Output: photo_clean.png
./gemini-watermark-remover photo.png

# Input: photo.jpg with custom suffix
# Output: photo_restored.jpg
./gemini-watermark-remover -s "_restored" photo.jpg

# Process entire folder
# Input: ./generated/*.png
# Output: ./generated/*_clean.png
./gemini-watermark-remover ./generated/
```

## Supported Formats

- PNG (lossless)
- JPEG/JPG (95% quality on output)

## Project Structure

```
gemini-watermark-remover/
├── main.go                 # CLI entry point and file handling
├── go.mod                  # Go module definition
├── README.md               # This file
└── watermark/
    ├── doc.go              # Package documentation
    ├── assets.go           # Embedded reference watermark images
    ├── assets_test.go      # Tests for asset loading
    ├── alphamap.go         # Alpha map extraction from reference images
    ├── alphamap_test.go    # Tests for alpha map calculation
    ├── engine.go           # Core watermark removal algorithm
    ├── engine_test.go      # Tests for watermark removal engine
    └── assets/
        ├── bg_48.png       # 48x48 reference (watermark on black)
        └── bg_96.png       # 96x96 reference (watermark on black)
```

## Testing

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run tests with coverage report:

```bash
go test -cover ./...
```

Generate detailed coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## How the Reference Images Work

The `bg_48.png` and `bg_96.png` files are screenshots of the Gemini watermark applied to a pure black background. Since the watermark is white with varying transparency:

- On a black background, each pixel's brightness equals its alpha value
- We extract `max(R, G, B)` for each pixel and normalize to get alpha
- This alpha map is then used in the reverse blending formula

## Limitations

- Only works with unmodified Gemini watermarks in the expected position
- If the image has been cropped, resized, or the watermark area has been edited, removal may not work correctly
- Very dark images in the watermark region may show slight artifacts

## Credits

Algorithm based on [gemini-watermark-remover](https://github.com/journey-ad/gemini-watermark-remover) by journey-ad.

## License

MIT License - see LICENSE file for details.
