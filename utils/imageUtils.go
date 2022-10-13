package utils

import (
	"fmt"
	"image"
	"image/color"

	"golang.org/x/image/draw"
)

// HorizontalLine draws a horizontal line onto an image
func HorizontalLine(img *image.RGBA, xStart int, xEnd int, y int, c color.Color) {
	for i := xStart; i <= xEnd; i++ {
		img.Set(i, y, c)
	}
}

// VerticalLine draws a vertical line onto an image
func VerticalLine(img *image.RGBA, yStart int, yEnd int, x int, c color.Color) {
	for i := yStart; i <= yEnd; i++ {
		img.Set(x, i, c)
	}
}

// Rectangle draws a rectangle onto an image
func Rectangle(img *image.RGBA, xStart int, xEnd int, yStart int, yEnd int, c color.Color) {
	HorizontalLine(img, xStart, xEnd, yStart, c)
	HorizontalLine(img, xStart, xEnd, yEnd, c)
	VerticalLine(img, yStart, yEnd, xStart, c)
	VerticalLine(img, yStart, yEnd, xEnd, c)
}

// Scale scales a given image to the desired dimensions
func Scale(img *image.RGBA, xStart int, yStart int, xEnd int, yEnd int, interpolator draw.Interpolator) image.Image {
	originalBounds := img.Bounds()
	scaledImage := image.NewRGBA(image.Rect(xStart, yStart, xEnd, yEnd))

	// TODO: evaluate other algorithms
	interpolator.Scale(scaledImage, scaledImage.Bounds(), img, originalBounds, draw.Over, nil)
	return scaledImage
}

// ComparePixelsExact naively compares two images by checking how many of the pixels between them are identical.
// It returns a float that ranges between 0 (no matches) and 1 (identical pictures)
func ComparePixelsExact(imageA *image.RGBA, imageB *image.RGBA, globalBounds image.Rectangle) (float64, error) {
	// Ensure that images dimensions and origin points are equal
	if imageA.Bounds().Min.X != imageB.Bounds().Min.X ||
		imageA.Bounds().Min.Y != imageB.Bounds().Min.Y ||
		imageA.Bounds().Max.X != imageB.Bounds().Max.X ||
		imageA.Bounds().Max.Y != imageB.Bounds().Max.Y {
		return 0, fmt.Errorf("bounds for image A (%v) and image B (%v) do not match", imageA.Bounds(), imageB.Bounds())
	}

	matches := 0
	skipped := 0

	// Compare every pixel across both images
	for x := imageA.Bounds().Min.X; x < imageA.Bounds().Max.X; x++ {
		for y := imageA.Bounds().Min.Y; y < imageA.Bounds().Max.Y; y++ {

			// The padding is of no interest
			if OutOfBounds(image.Point{X: x, Y: y}, globalBounds) {
				skipped++
				continue
			}

			aR, aG, aB, _ := imageA.At(x, y).RGBA()
			bR, bG, bB, _ := imageB.At(x, y).RGBA()

			if aR == bR && aG == bG && aB == bB {
				matches++
			}
		}
	}

	// If none of the checked pixels were inside bounds, signal that no further partitioning is required
	relevantPixels := imageA.Bounds().Dx()*imageA.Bounds().Dy() - skipped
	if relevantPixels <= 0 {
		return 1, nil
	}

	// Else compute similarity without OOB pixels
	similarity := float64(matches) / float64(relevantPixels)
	return similarity, nil
}

// OutOfBounds returns whether a point is in a rectangle
func OutOfBounds(point image.Point, bounds image.Rectangle) bool {
	x := point.X
	y := point.Y

	rectXStart := bounds.Min.X
	rectYStart := bounds.Min.Y
	rectXEnd := bounds.Max.X
	rectYEnd := bounds.Max.Y

	return x < rectXStart ||
		x > rectXEnd ||
		y < rectYStart ||
		y > rectYEnd
}
