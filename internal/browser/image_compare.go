package browser

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// CompareImages compares two image byte arrays and returns a similarity score
// Returns a value between 0.0 (completely different) and 1.0 (identical)
func CompareImages(img1Bytes, img2Bytes []byte) (float64, error) {
	// Decode first image
	img1, err := png.Decode(bytes.NewReader(img1Bytes))
	if err != nil {
		return 0, fmt.Errorf("failed to decode first image: %w", err)
	}

	// Decode second image
	img2, err := png.Decode(bytes.NewReader(img2Bytes))
	if err != nil {
		return 0, fmt.Errorf("failed to decode second image: %w", err)
	}

	// Check if dimensions match
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	// If dimensions don't match, scale the larger image down to match the smaller one
	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		// Determine which image is larger
		if bounds1.Dx() > bounds2.Dx() || bounds1.Dy() > bounds2.Dy() {
			// Scale img1 down to match img2
			img1 = scaleImage(img1, bounds2.Dx(), bounds2.Dy())
			bounds1 = img1.Bounds()
		} else {
			// Scale img2 down to match img1
			img2 = scaleImage(img2, bounds1.Dx(), bounds1.Dy())
			bounds2 = img2.Bounds()
		}
	}

	// Calculate MSE (Mean Squared Error)
	var totalError float64
	pixelCount := bounds1.Dx() * bounds1.Dy()

	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
			r1, g1, b1, a1 := img1.At(x, y).RGBA()
			r2, g2, b2, a2 := img2.At(x, y).RGBA()

			// Convert from uint32 (0-65535) to float64 (0-255)
			dr := float64(r1>>8) - float64(r2>>8)
			dg := float64(g1>>8) - float64(g2>>8)
			db := float64(b1>>8) - float64(b2>>8)
			da := float64(a1>>8) - float64(a2>>8)

			// Sum of squared differences for all channels
			totalError += dr*dr + dg*dg + db*db + da*da
		}
	}

	// Calculate MSE
	mse := totalError / float64(pixelCount*4) // 4 channels (RGBA)

	// Convert MSE to similarity score (0-1)
	// MSE ranges from 0 (identical) to 255^2 (completely different)
	// We invert and normalize it to get similarity
	maxMSE := 255.0 * 255.0
	similarity := 1.0 - math.Min(mse/maxMSE, 1.0)

	return similarity, nil
}

// PixelDifferenceCount counts how many pixels are different between two images
func PixelDifferenceCount(img1Bytes, img2Bytes []byte, threshold uint32) (int, error) {
	// Decode images
	img1, err := png.Decode(bytes.NewReader(img1Bytes))
	if err != nil {
		return 0, fmt.Errorf("failed to decode first image: %w", err)
	}

	img2, err := png.Decode(bytes.NewReader(img2Bytes))
	if err != nil {
		return 0, fmt.Errorf("failed to decode second image: %w", err)
	}

	// Check dimensions
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	// If dimensions don't match, scale the larger image down to match the smaller one
	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		if bounds1.Dx() > bounds2.Dx() || bounds1.Dy() > bounds2.Dy() {
			img1 = scaleImage(img1, bounds2.Dx(), bounds2.Dy())
			bounds1 = img1.Bounds()
		} else {
			img2 = scaleImage(img2, bounds1.Dx(), bounds1.Dy())
			bounds2 = img2.Bounds()
		}
	}

	// Count different pixels
	differentPixels := 0

	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
			r1, g1, b1, a1 := img1.At(x, y).RGBA()
			r2, g2, b2, a2 := img2.At(x, y).RGBA()

			// Calculate difference in each channel
			dr := int32(r1) - int32(r2)
			dg := int32(g1) - int32(g2)
			db := int32(b1) - int32(b2)
			da := int32(a1) - int32(a2)

			// Check if any channel differs by more than threshold
			if abs32(dr) > int32(threshold) ||
				abs32(dg) > int32(threshold) ||
				abs32(db) > int32(threshold) ||
				abs32(da) > int32(threshold) {
				differentPixels++
			}
		}
	}

	return differentPixels, nil
}

func abs32(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}

// scaleImage scales an image to the target width and height using nearest neighbor
func scaleImage(src image.Image, targetWidth, targetHeight int) image.Image {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	// Use nearest neighbor scaling
	xRatio := float64(srcWidth) / float64(targetWidth)
	yRatio := float64(srcHeight) / float64(targetHeight)

	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			srcX := int(float64(x) * xRatio)
			srcY := int(float64(y) * yRatio)
			dst.Set(x, y, src.At(srcBounds.Min.X+srcX, srcBounds.Min.Y+srcY))
		}
	}

	return dst
}

// CreateDiffImage creates a visual diff image highlighting differences between two images
// Identical pixels are shown in grayscale, different pixels are highlighted in red
// Returns the diff image as PNG bytes, and optionally saves to filePath if provided
func CreateDiffImage(img1Bytes, img2Bytes []byte, filePath string) ([]byte, error) {
	// Decode first image
	img1, err := png.Decode(bytes.NewReader(img1Bytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode first image: %w", err)
	}

	// Decode second image
	img2, err := png.Decode(bytes.NewReader(img2Bytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode second image: %w", err)
	}

	// Check if dimensions match
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	// If dimensions don't match, scale the larger image down to match the smaller one
	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		if bounds1.Dx() > bounds2.Dx() || bounds1.Dy() > bounds2.Dy() {
			img1 = scaleImage(img1, bounds2.Dx(), bounds2.Dy())
			bounds1 = img1.Bounds()
		} else {
			img2 = scaleImage(img2, bounds1.Dx(), bounds1.Dy())
			bounds2 = img2.Bounds()
		}
	}

	// Create diff image
	width := bounds1.Dx()
	height := bounds1.Dy()
	diffImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// Threshold for considering pixels different (adjust as needed)
	const threshold = 10

	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
			r1, g1, b1, a1 := img1.At(x, y).RGBA()
			r2, g2, b2, a2 := img2.At(x, y).RGBA()

			// Convert from uint32 (0-65535) to uint8 (0-255)
			r1b := uint8(r1 >> 8)
			g1b := uint8(g1 >> 8)
			b1b := uint8(b1 >> 8)
			a1b := uint8(a1 >> 8)

			r2b := uint8(r2 >> 8)
			g2b := uint8(g2 >> 8)
			b2b := uint8(b2 >> 8)
			a2b := uint8(a2 >> 8)

			// Calculate difference
			dr := abs(int(r1b) - int(r2b))
			dg := abs(int(g1b) - int(g2b))
			db := abs(int(b1b) - int(b2b))
			da := abs(int(a1b) - int(a2b))

			// Check if pixels are different
			if dr > threshold || dg > threshold || db > threshold || da > threshold {
				// Highlight difference in red
				diffImg.SetRGBA(x-bounds1.Min.X, y-bounds1.Min.Y, color.RGBA{
					R: 255,
					G: 0,
					B: 0,
					A: 255,
				})
			} else {
				// Show identical pixels in grayscale (average of RGB)
				gray := uint8((int(r1b) + int(g1b) + int(b1b)) / 3)
				diffImg.SetRGBA(x-bounds1.Min.X, y-bounds1.Min.Y, color.RGBA{
					R: gray,
					G: gray,
					B: gray,
					A: a1b,
				})
			}
		}
	}

	// Encode diff image to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, diffImg); err != nil {
		return nil, fmt.Errorf("failed to encode diff image: %w", err)
	}

	diffBytes := buf.Bytes()

	// Save to file if path provided
	if filePath != "" {
		if err := os.WriteFile(filePath, diffBytes, 0644); err != nil {
			return nil, fmt.Errorf("failed to write diff image to %s: %w", filePath, err)
		}
	}

	return diffBytes, nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
