package quadtreeImage

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/xaverhimmelsbach/quadtree-block-compression/utils"
)

type QuadtreeImage struct {
	BaseImage   image.Image
	PaddedImage image.Image
	Child       *QuadtreeElement
}

// Partition splits the BaseImage into an appropriate number of sub images and calls their partition method
func (q *QuadtreeImage) Partition() {
	q.pad()

	childImage := image.NewRGBA(image.Rect(0, 0, q.PaddedImage.Bounds().Max.X-1, q.PaddedImage.Bounds().Max.Y-1))
	draw.Draw(childImage, childImage.Bounds(), q.PaddedImage, q.PaddedImage.Bounds().Min, draw.Src)

	q.Child = &QuadtreeElement{BaseImage: childImage}
	q.Child.partition()
}

// TODO: Implement
func (q *QuadtreeImage) Encode() {
	fmt.Println("Encoding QuadtreeImage")
}

// TODO: Implement
func (q *QuadtreeImage) WriteFile(path string) {
	fmt.Printf("Writing QuadtreeImage to %q\n", path)
}

// Visualize draws the bounding boxes of all Children onto a copy of the BaseImage and of the PaddedImage
func (q *QuadtreeImage) Visualize(path string) (image.Image, image.Image, error) {
	rects := q.Child.visualize()
	baseBounds := q.BaseImage.Bounds()
	paddedBounds := q.PaddedImage.Bounds()

	baseImage := image.NewRGBA(image.Rect(0, 0, baseBounds.Dx(), baseBounds.Dy()))
	draw.Draw(baseImage, baseImage.Bounds(), q.BaseImage, baseBounds.Min, draw.Src)

	paddedImage := image.NewRGBA(image.Rect(0, 0, paddedBounds.Dx(), paddedBounds.Dy()))
	draw.Draw(paddedImage, paddedImage.Bounds(), q.PaddedImage, paddedBounds.Min, draw.Src)

	for _, rect := range rects {
		utils.Rectangle(baseImage, rect.Min.X, rect.Max.X, rect.Min.Y, rect.Max.Y, color.RGBA{R: 255, A: 255})
		utils.Rectangle(paddedImage, rect.Min.X, rect.Max.X, rect.Min.Y, rect.Max.Y, color.RGBA{R: 255, A: 255})
	}

	return baseImage, paddedImage, nil
}

// pad adds transparent padding to a copy of BaseImage to make it a square with an edge length that can be divided by a multiple of four to get a JPEG block
func (q *QuadtreeImage) pad() {
	baseBounds := q.BaseImage.Bounds()
	var longerSideLength int
	paddedSideLength := 8

	// Find the longer side of X and Y
	if baseBounds.Dx() > baseBounds.Dy() {
		longerSideLength = baseBounds.Dx()
	} else {
		longerSideLength = baseBounds.Dy()
	}

	// Pad until the padding is greater than both sides of the BaseImage
	for paddedSideLength < longerSideLength {
		paddedSideLength *= 4
	}

	// Copy BaseImage over padded image
	paddedImage := image.NewRGBA(image.Rect(0, 0, paddedSideLength, paddedSideLength))
	draw.Draw(paddedImage, paddedImage.Bounds(), q.BaseImage, q.BaseImage.Bounds().Min, draw.Src)

	q.PaddedImage = paddedImage
}
