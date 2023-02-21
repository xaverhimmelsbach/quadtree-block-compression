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
func Rectangle(img *image.RGBA, rectangle image.Rectangle, c color.Color, fillColor color.Color) {
	draw.Draw(img, rectangle, &image.Uniform{fillColor}, rectangle.Min, draw.Over)
	HorizontalLine(img, rectangle.Min.X, rectangle.Max.X, rectangle.Min.Y, c)
	HorizontalLine(img, rectangle.Min.X, rectangle.Max.X, rectangle.Max.Y, c)
	VerticalLine(img, rectangle.Min.Y, rectangle.Max.Y, rectangle.Min.X, c)
	VerticalLine(img, rectangle.Min.Y, rectangle.Max.Y, rectangle.Max.X, c)
}

// Scale scales a given image to the desired dimensions
func Scale(img *image.RGBA, bounds image.Rectangle, interpolator draw.Interpolator) image.Image {
	originalBounds := img.Bounds()
	scaledImage := image.NewRGBA(bounds)

	// TODO: evaluate other algorithms
	interpolator.Scale(scaledImage, scaledImage.Bounds(), img, originalBounds, draw.Over, nil)
	return scaledImage
}

type FillConfig struct {
	// Does this edge/corner need to be filled towards the edge/corner of img?
	shouldFill bool
	// Bounding box of the edge/pixel to be used as a filling
	copyBounds image.Rectangle
	// Bounding box of the scaled filling
	scaleBounds image.Rectangle
}

// FillSpace fills everything in img that doesn't fall within nonTransparentImageBounds with the scaled out edges and corners of img when masked by nonTransparentImageBounds
func FillSpace(img *image.RGBA, nonTransparentImageBounds image.Rectangle) {
	var copyBaseImage *image.RGBA
	var scaledImage *image.RGBA

	shouldFillRight := nonTransparentImageBounds.Max.X < img.Bounds().Max.X
	shouldFillUpper := nonTransparentImageBounds.Min.Y > img.Bounds().Min.Y
	shouldFillLeft := nonTransparentImageBounds.Min.X > img.Bounds().Min.X
	shouldFillBottom := nonTransparentImageBounds.Max.Y < img.Bounds().Max.Y

	operations := []FillConfig{
		// Right edge
		{
			shouldFill:  shouldFillRight,
			copyBounds:  image.Rect(nonTransparentImageBounds.Max.X-1, nonTransparentImageBounds.Min.Y, nonTransparentImageBounds.Max.X, nonTransparentImageBounds.Max.Y),
			scaleBounds: image.Rect(nonTransparentImageBounds.Max.X, nonTransparentImageBounds.Min.Y, img.Bounds().Max.X, nonTransparentImageBounds.Max.Y),
		},
		// Upper edge
		{
			shouldFill:  shouldFillUpper,
			copyBounds:  image.Rect(nonTransparentImageBounds.Min.X, nonTransparentImageBounds.Min.Y, nonTransparentImageBounds.Max.X, nonTransparentImageBounds.Min.Y+1),
			scaleBounds: image.Rect(nonTransparentImageBounds.Min.X, img.Bounds().Min.Y, nonTransparentImageBounds.Max.X, nonTransparentImageBounds.Min.Y),
		},
		// Left edge
		{
			shouldFill:  shouldFillLeft,
			copyBounds:  image.Rect(nonTransparentImageBounds.Min.X, nonTransparentImageBounds.Min.Y, nonTransparentImageBounds.Min.X+1, nonTransparentImageBounds.Max.Y),
			scaleBounds: image.Rect(img.Bounds().Min.X, nonTransparentImageBounds.Min.Y, nonTransparentImageBounds.Min.X, nonTransparentImageBounds.Max.Y),
		},
		// Bottom edge
		{
			shouldFill:  shouldFillBottom,
			copyBounds:  image.Rect(nonTransparentImageBounds.Min.X, nonTransparentImageBounds.Max.Y-1, nonTransparentImageBounds.Max.X, nonTransparentImageBounds.Max.Y),
			scaleBounds: image.Rect(nonTransparentImageBounds.Min.X, nonTransparentImageBounds.Max.Y, nonTransparentImageBounds.Max.X, img.Bounds().Max.Y),
		},
		// Upper right corner
		{
			shouldFill:  shouldFillRight && shouldFillUpper,
			copyBounds:  image.Rect(nonTransparentImageBounds.Max.X, nonTransparentImageBounds.Min.Y, nonTransparentImageBounds.Max.X+1, nonTransparentImageBounds.Min.Y+1),
			scaleBounds: image.Rect(nonTransparentImageBounds.Max.X, img.Bounds().Min.Y, img.Bounds().Max.X, nonTransparentImageBounds.Min.Y),
		},
		// Upper left corner
		{
			shouldFill:  shouldFillLeft && shouldFillUpper,
			copyBounds:  image.Rect(nonTransparentImageBounds.Min.X, nonTransparentImageBounds.Min.Y, nonTransparentImageBounds.Min.X-1, nonTransparentImageBounds.Min.Y-1),
			scaleBounds: image.Rect(img.Bounds().Min.X, img.Bounds().Min.Y, nonTransparentImageBounds.Min.X, nonTransparentImageBounds.Min.Y),
		},
		// Bottom left corner
		{
			shouldFill:  shouldFillLeft && shouldFillBottom,
			copyBounds:  image.Rect(nonTransparentImageBounds.Min.X, nonTransparentImageBounds.Max.Y-1, nonTransparentImageBounds.Min.X+1, nonTransparentImageBounds.Max.Y),
			scaleBounds: image.Rect(img.Bounds().Min.X, nonTransparentImageBounds.Max.Y, nonTransparentImageBounds.Min.X, img.Bounds().Max.Y),
		},
		// Bottom right corner
		{
			shouldFill:  shouldFillRight && shouldFillBottom,
			copyBounds:  image.Rect(nonTransparentImageBounds.Max.X-1, nonTransparentImageBounds.Max.Y-1, nonTransparentImageBounds.Max.X, nonTransparentImageBounds.Max.Y),
			scaleBounds: image.Rect(nonTransparentImageBounds.Max.X, nonTransparentImageBounds.Max.Y, img.Bounds().Max.X, img.Bounds().Max.Y),
		},
	}

	// TODO: Prime candidate for multithreading
	for _, operation := range operations {
		// If the current edge doesn't lie right at the edge of img
		if operation.shouldFill {
			// Get the current edge or edge point
			copyBaseImage = image.NewRGBA(operation.copyBounds)
			draw.Draw(copyBaseImage, copyBaseImage.Bounds(), img, copyBaseImage.Bounds().Min, draw.Over)

			// Scale it up towards the img edge
			scaledImage = Scale(copyBaseImage, operation.scaleBounds, draw.NearestNeighbor).(*image.RGBA)

			// Copy the scaled image into img
			draw.Draw(img, operation.scaleBounds, scaledImage, scaledImage.Bounds().Min, draw.Over)
		}
	}
}
