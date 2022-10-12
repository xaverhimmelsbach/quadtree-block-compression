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

// Partition splits the BaseImage into an appropriate number of sub images and calls their partition method
func (q *QuadtreeImage) Partition() {
	// TODO: create more than one child
	childImage := image.NewRGBA(image.Rect(0, 0, q.BaseImage.Bounds().Max.X-1, q.BaseImage.Bounds().Max.Y-1))
	draw.Draw(childImage, childImage.Bounds(), q.BaseImage, q.BaseImage.Bounds().Min, draw.Src)

	q.Children = make([]*QuadtreeElement, 0)
	q.Children = append(q.Children, &QuadtreeElement{BaseImage: childImage})

	for _, child := range q.Children {
		child.partition()
	}
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
