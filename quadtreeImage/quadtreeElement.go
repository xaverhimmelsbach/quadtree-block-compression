package quadtreeImage

import (
	"fmt"
	"image"
)

type QuadtreeElement struct {
	BaseImage        image.Image
	DownsampledImage image.Image
	Children         []*QuadtreeElement
}

// TODO: Implement
func (q *QuadtreeElement) partition() {
	fmt.Println("Partitioning QuadtreeElement")
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
