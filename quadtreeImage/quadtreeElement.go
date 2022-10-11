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
func (*QuadtreeElement) partition() {
	fmt.Println("Partitioning QuadtreeElement")
}

// TODO: Implement
func (*QuadtreeElement) createDownsampledImage() {
	fmt.Println("Creating Downsampled Image")
}

// TODO: Implement
func (*QuadtreeElement) compareImages() {
	fmt.Println("Comparing Images")
}
