package utils

import (
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

// Scale scales a given image to the desired width and height
func Scale(img *image.RGBA, desiredWidth int, desiredHeight int) image.Image {
	originalBounds := img.Bounds()
	scaledImage := image.NewRGBA(image.Rect(0, 0, desiredWidth, desiredHeight))

	// TODO: evaluate other algorithms
	draw.NearestNeighbor.Scale(scaledImage, scaledImage.Bounds(), img, originalBounds, draw.Over, nil)
	return scaledImage
}
