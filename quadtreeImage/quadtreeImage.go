package quadtreeImage

import (
	"fmt"
	"image"
)

type QuadtreeImage struct {
	BaseImage image.Image
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
