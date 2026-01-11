# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-01-12

### Added
- Support for glob patterns (`"*.png"`, `"photos/*.jpg"`)
- Support for multiple file arguments (`img1.png img2.jpg img3.png`)
- Ability to mix files, directories, and globs in a single command
- Skip files that already have the output suffix in all input modes
- New tests for glob and multiple file functionality

## [0.1.0] - 2026-01-12

### Added
- Initial release
- Remove Gemini AI watermarks from PNG and JPEG images
- Automatic watermark size detection (48x48 or 96x96 based on image dimensions)
- Single file processing
- Batch processing of directories
- Customizable output suffix (`-s` flag)
- Verbose mode (`-v` flag) for detailed processing information
- Quiet mode (`-q` flag) for minimal output
- Preserves original image format (PNG stays PNG, JPEG stays JPEG)
- Embedded reference watermark images (no external dependencies)
- Cross-platform binaries (Linux, macOS, Windows)

### Technical Details
- Uses reverse alpha blending algorithm to restore original pixels
- Reference images capture watermark applied to black background
- Alpha map extracted from max(R,G,B) of each pixel
- Formula: `original = (watermarked - alpha * 255) / (1 - alpha)`

[Unreleased]: https://github.com/cnu/gemini-watermark-remover/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/cnu/gemini-watermark-remover/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/cnu/gemini-watermark-remover/releases/tag/v0.1.0
