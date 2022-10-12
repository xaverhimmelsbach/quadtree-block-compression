package utils

import (
	"image"
	"image/color"
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
