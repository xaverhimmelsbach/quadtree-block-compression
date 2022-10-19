package quadtreeImage

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/xaverhimmelsbach/quadtree-block-compression/config"
	"github.com/xaverhimmelsbach/quadtree-block-compression/utils"
)

type QuadtreeImage struct {
	baseImage   image.Image
	paddedImage image.Image
	child       *QuadtreeElement
	config      *config.Config
}

// NewQuadtreeImage constructs a well-formed instance of QuadtreeImage from a baseImage
func NewQuadtreeImage(baseImage image.Image, cfg *config.Config) *QuadtreeImage {
	qti := new(QuadtreeImage)

	qti.config = cfg
	qti.baseImage = baseImage
	qti.paddedImage = qti.pad()

	return qti
}

// Partition splits the BaseImage into an appropriate number of sub images and calls their partition method
func (q *QuadtreeImage) Partition() {
	// Create root of the quadtree
	childImage := image.NewRGBA(image.Rect(0, 0, q.paddedImage.Bounds().Max.X, q.paddedImage.Bounds().Max.Y))
	draw.Draw(childImage, childImage.Bounds(), q.paddedImage, q.paddedImage.Bounds().Min, draw.Src)
	q.child = NewQuadtreeElement("", childImage, q.baseImage.Bounds(), q.config)

	// Start partitioning the quadtree
	q.child.partition()
}

// TODO: Implement
func (q *QuadtreeImage) Encode() {
	fmt.Println("Encoding QuadtreeImage")
}

// Visualize draws the bounding boxes of all Children onto a copy of the BaseImage and of the PaddedImage.
// It also draws them onto the upsampled JPEG blocks to show how the encoded result would look
func (q *QuadtreeImage) Visualize(path string, drawGrid bool) (image.Image, image.Image, image.Image, image.Image, error) {
	images := q.child.visualize()
	baseBounds := q.baseImage.Bounds()
	paddedBounds := q.paddedImage.Bounds()

	baseImage := image.NewRGBA(image.Rect(0, 0, baseBounds.Dx(), baseBounds.Dy()))
	draw.Draw(baseImage, baseImage.Bounds(), q.baseImage, baseBounds.Min, draw.Src)

	paddedImage := image.NewRGBA(image.Rect(0, 0, paddedBounds.Dx(), paddedBounds.Dy()))
	draw.Draw(paddedImage, paddedImage.Bounds(), q.paddedImage, paddedBounds.Min, draw.Src)

	baseImageBlocks := image.NewRGBA(image.Rect(0, 0, baseBounds.Dx(), baseBounds.Dy()))

	paddedImageBlocks := image.NewRGBA(image.Rect(0, 0, paddedBounds.Dx(), paddedBounds.Dy()))

	for _, img := range images {
		// Draw bounding boxes
		if drawGrid {
			utils.Rectangle(baseImage, img.Bounds().Min.X, img.Bounds().Max.X, img.Bounds().Min.Y, img.Bounds().Max.Y, color.RGBA{R: 255, A: 255})
			utils.Rectangle(paddedImage, img.Bounds().Min.X, img.Bounds().Max.X, img.Bounds().Min.Y, img.Bounds().Max.Y, color.RGBA{R: 255, A: 255})
		}

		// Combine separate upscaled blocks into whole images
		draw.Draw(baseImageBlocks, baseImageBlocks.Bounds(), img, img.Bounds().Min, draw.Src)
		draw.Draw(paddedImageBlocks, paddedImageBlocks.Bounds(), img, img.Bounds().Min, draw.Src)
	}

	// Additional loop to draw bounding boxes on top of the block images
	if drawGrid {
		for _, img := range images {
			utils.Rectangle(baseImageBlocks, img.Bounds().Min.X, img.Bounds().Max.X, img.Bounds().Min.Y, img.Bounds().Max.Y, color.RGBA{R: 255, A: 255})
			utils.Rectangle(paddedImageBlocks, img.Bounds().Min.X, img.Bounds().Max.X, img.Bounds().Min.Y, img.Bounds().Max.Y, color.RGBA{R: 255, A: 255})
		}
	}

	return baseImage, paddedImage, baseImageBlocks, paddedImageBlocks, nil
}

// pad adds transparent padding to a copy of BaseImage to make it a square with an edge length that can be divided by a multiple of four to get a JPEG block
func (q *QuadtreeImage) pad() image.Image {
	baseBounds := q.baseImage.Bounds()
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
	draw.Draw(paddedImage, paddedImage.Bounds(), q.baseImage, q.baseImage.Bounds().Min, draw.Src)

	utils.FillSpace(paddedImage, q.baseImage.Bounds())

	return paddedImage
}
