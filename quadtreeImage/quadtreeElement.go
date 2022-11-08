package quadtreeImage

import (
	"archive/zip"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"strconv"
	"strings"

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
	baseImage      image.Image
	blockImage     image.Image
	blockImageJPEG image.Image
	children       []*QuadtreeElement
	globalBounds   image.Rectangle
	isLeaf         bool
	config         *config.Config
	id             string
}

func NewQuadtreeElement(id string, baseImage image.Image, globalBounds image.Rectangle, cfg *config.Config) *QuadtreeElement {
	qte := new(QuadtreeElement)

	qte.id = id
	qte.config = cfg
	qte.baseImage = baseImage
	qte.globalBounds = globalBounds
	qte.blockImage, qte.blockImageJPEG = qte.createBlockImage()
	qte.isLeaf = qte.checkIsLeaf()

	return qte
}

// partition splits the BaseImage into four sub images, if further partitioning is necessary and calls their partition methods
func (q *QuadtreeElement) partition() {
	q.children = make([]*QuadtreeElement, 0)

	if !q.isLeaf {
		// Partition BaseImage into sub images
		for i := 0; i < ChildCount; i++ {
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
			draw.Draw(childImage, childImage.Bounds(), q.baseImage, childImage.Bounds().Min, draw.Src)

			// Create and partition child
			child := NewQuadtreeElement(q.id+strconv.Itoa(i), childImage, q.globalBounds, q.config)
			q.children = append(q.children, child)
			child.partition()
		}
	}
}

// checkIsLeaf checks whether the current block needs to be partitioned further
func (q *QuadtreeElement) checkIsLeaf() bool {
	// If the size of a JPEG block was reached, don't partition further
	if q.baseImage.Bounds().Dx() <= BlockSize || q.baseImage.Bounds().Dy() <= BlockSize {
		return true
	}

	// All blocks with a similarity of less than this need to be split further
	cutoff := q.config.Quadtree.SimilarityCutoff

	return q.compareImages() > cutoff
}

// createBlockImage scales the baseImage down to the size of a JPEG block and then scales it back up to the original size
func (q *QuadtreeElement) createBlockImage() (image.Image, image.Image) {
	baseImage := q.baseImage.(*image.RGBA)

	downsamplingInterpolator, err := getInterpolator(q.config.Quadtree.DownsamplingInterpolator)
	if err != nil {
		panic(err)
	}
	downsampledImage := utils.Scale(baseImage, image.Rect(0, 0, BlockSize, BlockSize), downsamplingInterpolator)
	downsampledImageRGBA := downsampledImage.(*image.RGBA)

	upsamplingInterpolator, err := getInterpolator(q.config.Quadtree.UpsamplingInterpolator)
	if err != nil {
		panic(err)
	}
	blockImage := utils.Scale(downsampledImageRGBA,
		image.Rect(q.baseImage.Bounds().Min.X, q.baseImage.Bounds().Min.Y, q.baseImage.Bounds().Max.X, q.baseImage.Bounds().Max.Y),
		upsamplingInterpolator).(*image.RGBA)

	return blockImage, downsampledImage
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

// encode writes the quadtree structure to a zip file
func (q *QuadtreeElement) encode(zipWriter *zip.Writer) (err error) {
	// Create directory path in zip file
	// TODO: can this be optimized?
	path := strings.Join(strings.Split(q.id, ""), "/")

	if q.isLeaf {
		// Either create and encode an image file if this is a quadtree leaf
		fileWriter, err := zipWriter.Create(path)
		if err != nil {
			return err
		}

		err = jpeg.Encode(fileWriter, q.blockImageJPEG, nil)
		if err != nil {
			return err
		}
	} else {
		// Or recurse into children
		for _, child := range q.children {
			child.encode(zipWriter)
		}
	}

	return nil
}

func (q *QuadtreeElement) decode(path string, file *zip.File, remainingHeight int) error {
	if path == "" {
		// scale up file if needed and use as blockImage
		fmt.Printf("Finished decoding at height %d\n", remainingHeight)

		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		fileImage, err := utils.ReadImageFromReader(fileReader)
		if err != nil {
			return err
		}
		fileImageRGBA := image.NewRGBA(fileImage.Bounds())
		draw.Draw(fileImageRGBA, fileImageRGBA.Bounds(), fileImage, fileImage.Bounds().Min, draw.Src)

		// Duplicate code fragment
		upsamplingInterpolator, err := getInterpolator(q.config.Quadtree.UpsamplingInterpolator)
		if err != nil {
			panic(err)
		}
		blockImage := utils.Scale(fileImageRGBA,
			image.Rect(q.baseImage.Bounds().Min.X, q.baseImage.Bounds().Min.Y, q.baseImage.Bounds().Max.X, q.baseImage.Bounds().Max.Y),
			upsamplingInterpolator).(*image.RGBA)

		q.blockImage = blockImage

		return nil
	}

	if remainingHeight == 0 {
		return fmt.Errorf("further partition according to path %s would lead to remaining height being smaller than 0 in %s", path, q.id)
	}

	if len(q.children) != ChildCount {
		// TODO: More or less duplicate code fragment
		for i := 0; i < ChildCount; i++ {
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

			// Create and partition child
			// NewQuadtreeElement(q.id+strconv.Itoa(i), childImage, q.globalBounds, q.config)
			child := &QuadtreeElement{
				id:        q.id + strconv.Itoa(i),
				baseImage: childImage,
				config:    q.config,
			}
			q.children = append(q.children, child)
		}
	}

	splitPath := strings.Split(path, "/")
	childId, err := strconv.Atoi(splitPath[0])
	if err != nil {
		return err
	}

	if childId >= ChildCount {
		return fmt.Errorf("childId %d is greater than child count (%d)", childId, ChildCount)
	}

	recursePath := strings.Join(splitPath[1:], "/")

	return q.children[childId].decode(recursePath, file, remainingHeight-1)
}

// visualize returns its own bounding box if it has no children, else it returns its childrens bounding boxes
func (q *QuadtreeElement) visualize() []image.Image {
	rects := make([]image.Image, 0)

	if len(q.children) == 0 {
		rects = append(rects, q.blockImage)
	} else {
		for _, child := range q.children {
			rects = append(rects, child.visualize()...)
		}
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
