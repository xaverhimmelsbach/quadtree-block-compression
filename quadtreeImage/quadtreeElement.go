package quadtreeImage

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/xaverhimmelsbach/quadtree-block-compression/config"
	"github.com/xaverhimmelsbach/quadtree-block-compression/utils"
	drawX "golang.org/x/image/draw"
)

var interpolators = map[string]drawX.Interpolator{
	"NearestNeighbor": drawX.NearestNeighbor,
	"ApproxBilinear":  drawX.ApproxBiLinear,
	"Bilinear":        drawX.BiLinear,
	"CatmullRom":      drawX.CatmullRom,
}

type QuadtreeElement struct {
	baseImage    image.Image
	blockImage   image.Image
	children     []*QuadtreeElement
	globalBounds image.Rectangle
	isLeaf       bool
	config       *config.Config
}

func NewQuadtreeElement(baseImage image.Image, globalBounds image.Rectangle, cfg *config.Config) *QuadtreeElement {
	qte := new(QuadtreeElement)

	qte.config = cfg
	qte.baseImage = baseImage
	qte.globalBounds = globalBounds
	qte.blockImage = qte.createBlockImage()
	qte.isLeaf = qte.checkIsLeaf()

	return qte
}

// partition splits the BaseImage into four sub images, if further partitioning is necessary and calls their partition methods
func (q *QuadtreeElement) partition() {
	q.children = make([]*QuadtreeElement, 0)

	if !q.isLeaf {
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
			child := NewQuadtreeElement(childImage, q.globalBounds, q.config)
			q.children = append(q.children, child)
			child.partition()
		}
	}
}

// checkIsLeaf checks whether the current block needs to be partitioned further
func (q *QuadtreeElement) checkIsLeaf() bool {
	// If the size of a JPEG block was reached, don't partition further
	if q.baseImage.Bounds().Dx() <= 8 || q.baseImage.Bounds().Dy() <= 8 {
		return true
	}

	// All blocks with a similarity of less than this need to be split further
	cutoff := q.config.Quadtree.SimilarityCutoff

	return q.compareImages() > cutoff
}

// createBlockImage scales the baseImage down to the size of a JPEG block and then scales it back up to the original size
func (q *QuadtreeElement) createBlockImage() image.Image {
	baseImage := q.baseImage.(*image.RGBA)

	downsamplingInterpolator, err := getInterpolator(q.config.Quadtree.DownsamplingInterpolator)
	if err != nil {
		panic(err)
	}
	downsampledImage := utils.Scale(baseImage, 0, 0, 8, 8, downsamplingInterpolator).(*image.RGBA)

	upsamplingInterpolator, err := getInterpolator(q.config.Quadtree.UpsamplingInterpolator)
	if err != nil {
		panic(err)
	}
	blockImage := utils.Scale(downsampledImage,
		q.baseImage.Bounds().Min.X, q.baseImage.Bounds().Min.Y,
		q.baseImage.Bounds().Max.X, q.baseImage.Bounds().Max.Y,
		upsamplingInterpolator).(*image.RGBA)

	return blockImage
}

// compareImages compares the scaled down JPEG block with the base image of this element
func (q *QuadtreeElement) compareImages() float64 {
	baseImage := q.baseImage.(*image.RGBA)
	blockImage := q.blockImage.(*image.RGBA)

	similarity, err := utils.ComparePixelsWeighted(blockImage, baseImage, q.globalBounds)
	// TODO: Handle errors better
	if err != nil {
		panic(err)
	}

	return similarity
}

// visualize returns its own bounding box if it has no children, else it returns its childrens bounding boxes
func (q *QuadtreeElement) visualize() []image.Image {
	rects := make([]image.Image, 0)

	if len(q.children) == 0 {
		rects = append(rects, q.blockImage)
	} else {
		rects = append(rects, q.children[0].visualize()...)
		rects = append(rects, q.children[1].visualize()...)
		rects = append(rects, q.children[2].visualize()...)
		rects = append(rects, q.children[3].visualize()...)
	}

	return rects
}

func getInterpolator(interpolatorId string) (drawX.Interpolator, error) {
	interpolator, ok := interpolators[interpolatorId]
	var err error
	if !ok {
		err = fmt.Errorf("interpolator id not found: %q", interpolatorId)
	}
	return interpolator, err
}
