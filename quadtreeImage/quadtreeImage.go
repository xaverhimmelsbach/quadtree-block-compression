package quadtreeImage

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/xaverhimmelsbach/quadtree-block-compression/utils"
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

// Visualize draws the bounding boxes of all Children onto a copy of the BaseImage
func (q *QuadtreeImage) Visualize(path string) (image.Image, error) {
	rects := make([]image.Rectangle, 0)
	for _, child := range q.Children {
		rects = append(rects, child.visualize()...)
	}

	b := q.BaseImage.Bounds()
	img := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(img, img.Bounds(), q.BaseImage, b.Min, draw.Src)

	for _, rect := range rects {
		utils.Rectangle(img, rect.Min.X, rect.Max.X, rect.Min.Y, rect.Max.Y, color.RGBA{R: 255})
	}

	return img, nil
}
