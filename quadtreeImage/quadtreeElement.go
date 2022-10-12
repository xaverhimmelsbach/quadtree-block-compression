package quadtreeImage

import (
	"fmt"
	"image"
	"image/draw"
)

type QuadtreeElement struct {
	BaseImage        image.Image
	DownsampledImage image.Image
	Children         []*QuadtreeElement
}

// partition splits the BaseImage into four sub images, if further partitioning is necessary and calls their partition methods
func (q *QuadtreeElement) partition() {
	q.Children = make([]*QuadtreeElement, 0)

	if q.furtherPartitioningNecessary() {
		// Partition BaseImage into 4 sub images
		for i := 0; i < 4; i++ {
			var xStart int
			var yStart int
			var xEnd int
			var yEnd int

			// Set x coordinates
			if i&1 == 0 {
				xStart = q.BaseImage.Bounds().Min.X
				xEnd = q.BaseImage.Bounds().Min.X + q.BaseImage.Bounds().Dx()/2
			} else {
				xStart = q.BaseImage.Bounds().Min.X + q.BaseImage.Bounds().Dx()/2
				xEnd = q.BaseImage.Bounds().Max.X
			}

			// Set y coordinates
			if i&2 == 0 {
				yStart = q.BaseImage.Bounds().Min.Y
				yEnd = q.BaseImage.Bounds().Min.Y + q.BaseImage.Bounds().Dy()/2
			} else {
				yStart = q.BaseImage.Bounds().Min.Y + q.BaseImage.Bounds().Dy()/2
				yEnd = q.BaseImage.Bounds().Max.Y
			}

			// Copy BaseImage section to sub image
			img := image.NewRGBA(image.Rect(xStart, yStart, xEnd, yEnd))
			draw.Draw(img, img.Bounds(), q.BaseImage, q.BaseImage.Bounds().Min, draw.Src)

			// Create and partition child
			child := &QuadtreeElement{BaseImage: img}
			q.Children = append(q.Children, child)
			child.partition()
		}
	}
}

// TODO: Placeholder condition
func (q *QuadtreeElement) furtherPartitioningNecessary() bool {
	return q.BaseImage.Bounds().Dx() > 200 || q.BaseImage.Bounds().Dy() > 200
}

// TODO: Implement
func (q *QuadtreeElement) createDownsampledImage() {
	fmt.Println("Creating Downsampled Image")
}

// TODO: Implement
func (q *QuadtreeElement) compareImages() {
	fmt.Println("Comparing Images")
}

// visualize returns its own bounding box if it has no children, else it returns its childrens bounding boxes
func (q *QuadtreeElement) visualize() []image.Rectangle {
	rects := make([]image.Rectangle, 0)

	if len(q.Children) == 0 {
		rects = append(rects, q.BaseImage.Bounds())
	} else {
		rects = append(rects, q.Children[0].visualize()...)
		rects = append(rects, q.Children[1].visualize()...)
		rects = append(rects, q.Children[2].visualize()...)
		rects = append(rects, q.Children[3].visualize()...)
	}

	return rects
}
