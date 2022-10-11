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
