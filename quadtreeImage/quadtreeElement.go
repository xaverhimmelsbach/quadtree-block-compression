package quadtreeImage

import (
	"image"
	"image/draw"

	"github.com/xaverhimmelsbach/quadtree-block-compression/utils"
	drawX "golang.org/x/image/draw"
)

type QuadtreeElement struct {
	baseImage        image.Image
	downsampledImage image.Image
	children         []*QuadtreeElement
	globalBounds     image.Rectangle
}

// partition splits the BaseImage into four sub images, if further partitioning is necessary and calls their partition methods
func (q *QuadtreeElement) partition(baseImage image.Image, globalBounds image.Rectangle) {
	q.baseImage = baseImage
	q.globalBounds = globalBounds
	q.children = make([]*QuadtreeElement, 0)

	q.createDownsampledImage()

	if q.furtherPartitioningNecessary() {
		// Partition BaseImage into 4 sub images
		for i := 0; i < 4; i++ {
			var xStart int
			var yStart int
			var xEnd int
			var yEnd int

			// Set x coordinates
			if i&1 == 0 {
				xStart = q.baseImage.Bounds().Min.X
				xEnd = q.baseImage.Bounds().Min.X + q.baseImage.Bounds().Dx()/2
			} else {
				xStart = q.baseImage.Bounds().Min.X + q.baseImage.Bounds().Dx()/2
				xEnd = q.baseImage.Bounds().Max.X
			}

			// Set y coordinates
			if i&2 == 0 {
				yStart = q.baseImage.Bounds().Min.Y
				yEnd = q.baseImage.Bounds().Min.Y + q.baseImage.Bounds().Dy()/2
			} else {
				yStart = q.baseImage.Bounds().Min.Y + q.baseImage.Bounds().Dy()/2
				yEnd = q.baseImage.Bounds().Max.Y
			}

			// Copy BaseImage section to sub image
			childImage := image.NewRGBA(image.Rect(xStart, yStart, xEnd, yEnd))
			draw.Draw(childImage, childImage.Bounds(), q.baseImage, q.baseImage.Bounds().Min, draw.Src)

			// Create and partition child
			child := &QuadtreeElement{}
			q.children = append(q.children, child)
			child.partition(childImage, q.globalBounds)
		}
	}
}

// furtherPartitioningNecessary decides whether the current element needs to be split into four smaller ones.
// The decision is made upon the similarity of the JPEG block to the original base image
func (q *QuadtreeElement) furtherPartitioningNecessary() bool {
	// If the size of a JPEG block was reached, don't partition further
	if q.baseImage.Bounds().Dx() <= 8 || q.baseImage.Bounds().Dy() <= 8 {
		return false
	}

	// All blocks with a similarity of less than this need to be split further
	cutoff := 0.1

	return q.compareImages() < cutoff
}

// createDownsampledImage creates a representation of the base image that has been scaled down to the size of a JPEG block
func (q *QuadtreeElement) createDownsampledImage() {
	baseImage := q.baseImage.(*image.RGBA)
	downsampledImage := utils.Scale(baseImage, 0, 0, 8, 8, drawX.NearestNeighbor)
	q.downsampledImage = downsampledImage
}

// compareImages compares the scaled down JPEG block with the base image of this element
func (q *QuadtreeElement) compareImages() float64 {
	baseImage := q.baseImage.(*image.RGBA)
	baseBounds := baseImage.Bounds()

	downsampledImage := q.downsampledImage.(*image.RGBA)
	upsampledImage := utils.Scale(downsampledImage,
		baseBounds.Min.X, baseBounds.Min.Y,
		baseBounds.Max.X, baseBounds.Max.Y,
		drawX.NearestNeighbor).(*image.RGBA)

	similarity, err := utils.ComparePixels(upsampledImage, baseImage, q.globalBounds)
	// TODO: Handle errors better
	if err != nil {
		panic(err)
	}

	return similarity
}

// visualize returns its own bounding box if it has no children, else it returns its childrens bounding boxes
func (q *QuadtreeElement) visualize() []image.Rectangle {
	rects := make([]image.Rectangle, 0)

	if len(q.children) == 0 {
		rects = append(rects, q.baseImage.Bounds())
	} else {
		rects = append(rects, q.children[0].visualize()...)
		rects = append(rects, q.children[1].visualize()...)
		rects = append(rects, q.children[2].visualize()...)
		rects = append(rects, q.children[3].visualize()...)
	}

	return rects
}
