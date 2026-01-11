// Package watermark provides functionality to detect and remove Gemini AI watermarks
// from generated images using reverse alpha blending.
//
// # Overview
//
// Google's Gemini AI adds a semi-transparent white watermark to the bottom-right
// corner of generated images. This package removes the watermark by reversing
// the alpha compositing formula used to apply it.
//
// # Algorithm
//
// The watermark is applied using standard alpha blending:
//
//	watermarked = alpha * logo + (1 - alpha) * original
//
// Where:
//   - alpha is the transparency value (0.0 to 1.0)
//   - logo is 255 (white)
//   - original is the underlying pixel value
//
// We reverse this to recover the original:
//
//	original = (watermarked - alpha * 255) / (1 - alpha)
//
// # Watermark Detection
//
// The watermark size and position depend on the image dimensions:
//   - Images larger than 1024x1024: 96x96 watermark, 64px from edges
//   - Smaller images: 48x48 watermark, 32px from edges
//
// # Usage
//
// Basic usage:
//
//	engine, err := watermark.NewEngine()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Load your image
//	img, _, _ := image.Decode(file)
//
//	// Remove watermark
//	cleaned := engine.RemoveWatermark(img)
//
// # Reference Images
//
// The package embeds reference images (bg_48.png and bg_96.png) that contain
// the Gemini watermark applied to a black background. These are used to extract
// the alpha values for each pixel of the watermark.
package watermark
