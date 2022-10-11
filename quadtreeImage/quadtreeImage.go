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
func (q *QuadtreeImage) Partition() {
	fmt.Println("Partitioning QuadtreeImage")
}

// TODO: Implement
func (q *QuadtreeImage) Encode() {
	fmt.Println("Encoding QuadtreeImage")
}

// TODO: Implement
func (q *QuadtreeImage) WriteFile(path string) {
	fmt.Printf("Writing QuadtreeImage to %q\n", path)
}

// TODO: Implement
func (q *QuadtreeImage) Visualize(path string) {
	fmt.Printf("Visualizing QuadtreeImage to %q\n", path)
}
