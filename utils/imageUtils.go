package utils

import (
	"image"
	"image/color"

	"golang.org/x/image/draw"
)

const (
	weightedRed   float64 = 0.2989
	weightedGreen float64 = 0.5870
	weightedBlue  float64 = 0.1140
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
