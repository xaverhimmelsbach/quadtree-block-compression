package quadtreeImage

import (
	"fmt"
	"image"
)

type QuadtreeImage struct {
	BaseImage image.Image
	Children  []*QuadtreeElement
}

// TODO: Implement
func (*QuadtreeImage) Partition() {
	fmt.Println("Partitioning QuadtreeImage")
}

// TODO: Implement
func (*QuadtreeImage) Encode() {
	fmt.Println("Encoding QuadtreeImage")
}

// TODO: Implement
func (*QuadtreeImage) WriteFile(path string) {
	fmt.Printf("Writing QuadtreeImage to %q\n", path)
}

// TODO: Implement
func (*QuadtreeImage) Visualize(path string) {
	fmt.Printf("Visualizing QuadtreeImage to %q\n", path)
}
