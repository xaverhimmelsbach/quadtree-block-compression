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

// interpolators holds the different interpolating algorithms that can be used for scaling block images
var interpolators = map[string]drawX.Interpolator{
	"NearestNeighbor": drawX.NearestNeighbor,
	"ApproxBiLinear":  drawX.ApproxBiLinear,
	"BiLinear":        drawX.BiLinear,
	"CatmullRom":      drawX.CatmullRom,
}

// QuadtreeElement represents a node in the quadtree that can either be the parent of ChildCount children or contain a block image
type QuadtreeElement struct {
	// The section of the original image (with padding) that this QuadtreeElement occupies
	baseImage image.Image
	// baseImage scaled down to BlockSize
	blockImageMinimal image.Image
	// blockImageMinimal scaled back up to the size of baseImage
	blockImage image.Image
	// Children of this QuadtreeElement in the quadtree
	children []*QuadtreeElement
	// Bounding box of the original image, used for out-of-bounds-check
	globalBounds *image.Rectangle
	// Is this QuadtreeElement a leaf and does it therefore contain an actual blockImage?
	isLeaf bool
	// Can this block be skipped during encoding?
	canBeSkipped bool
	// Program configuration
	config *config.Config
	// Unique identifier of this QuadtreeElement
	id string
}

// VisualizationElement holds an image section and additional information relevant during visualization
type VisualizationElement struct {
	// Image section
	image        image.Image
	canBeSkipped bool
}

// NewQuadtreeElement returns a fully populated QuadtreeImage occupying the space of baseImage
func NewQuadtreeElement(id string, baseImage image.Image, globalBounds *image.Rectangle, cfg *config.Config) *QuadtreeElement {
	qte := new(QuadtreeElement)

	qte.id = id
	qte.config = cfg
	qte.baseImage = baseImage
	qte.globalBounds = globalBounds
	qte.blockImage, qte.blockImageMinimal = qte.createBlockImages()
	qte.isLeaf, qte.canBeSkipped = qte.checkIsLeaf()

	return qte
}

// partition splits the BaseImage into ChildCount subimages if further partitioning is required, and calls their partition methods
func (q *QuadtreeElement) partition() {
	q.children = make([]*QuadtreeElement, 0)

	if !q.isLeaf {
		// Partition BaseImage into sub images
		for i := 0; i < ChildCount; i++ {
			// TODO: this approach probably can't handle cases of ChildCount != 4
			var xStart, yStart, xEnd, yEnd int

			// Set x coordinates
			if i&1 == 0 {
				// Left block
				xStart = q.baseImage.Bounds().Min.X
				xEnd = q.baseImage.Bounds().Min.X + q.baseImage.Bounds().Dx()/2
			} else {
				// Right block
				xStart = q.baseImage.Bounds().Min.X + q.baseImage.Bounds().Dx()/2
				xEnd = q.baseImage.Bounds().Max.X
			}

			// Set y coordinates
			if i&2 == 0 {
				// Upper block
				yStart = q.baseImage.Bounds().Min.Y
				yEnd = q.baseImage.Bounds().Min.Y + q.baseImage.Bounds().Dy()/2
			} else {
				// Lower block
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

// checkIsLeaf checks whether the current block needs to be partitioned further and if it can be skipped during encoding
func (q *QuadtreeElement) checkIsLeaf() (bool, bool) {
	// If the current block is completely out of bounds it doesn't need further partitioning and can be skipped during encoding
	if !utils.RectanglesCollide(q.blockImage.Bounds(), *q.globalBounds) {
		return true, true
	}

	// If the minimal BlockSize was reached, don't partition further
	if q.baseImage.Bounds().Dx() <= BlockSize || q.baseImage.Bounds().Dy() <= BlockSize {
		return true, false
	}

	// Compare blockImage with baseImage
	return q.compareImages() > q.config.Quadtree.SimilarityCutoff, false
}

// createBlockImages scales the baseImage down to BlockSize and then scales it back up to the original size
func (q *QuadtreeElement) createBlockImages() (image.Image, image.Image) {
	baseImage := q.baseImage.(*image.RGBA)

	// Load inteprolators
	downsamplingInterpolator, err := getInterpolator(q.config.Quadtree.DownsamplingInterpolator)
	if err != nil {
		panic(err)
	}
	upsamplingInterpolator, err := getInterpolator(q.config.Quadtree.UpsamplingInterpolator)
	if err != nil {
		panic(err)
	}

	// Scale baseImage down to BlockSize
	downsampledImage := utils.Scale(baseImage, image.Rect(0, 0, BlockSize, BlockSize), downsamplingInterpolator)
	downsampledImageRGBA := downsampledImage.(*image.RGBA)

	// Scale downsampled image back up to size of baseImage
	blockImage := utils.Scale(downsampledImageRGBA, q.baseImage.Bounds(), upsamplingInterpolator).(*image.RGBA)

	return blockImage, downsampledImage
}

// compareImages compares blockImage with baseImage
func (q *QuadtreeElement) compareImages() float64 {
	baseImage := q.baseImage.(*image.RGBA)
	blockImage := q.blockImage.(*image.RGBA)

	similarity, err := utils.ComparePixelsWeighted(blockImage, baseImage, *q.globalBounds)
	// TODO: Handle errors better (e.g. by wrapping errors and returning them here as well)
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

	// Skip leaves that are out of bounds
	if q.isLeaf && (!q.config.Encoding.SkipOutOfBoundsBlocks || !q.canBeSkipped) {
		// Either create and encode an image file if this is a quadtree leaf
		fileWriter, err := zipWriter.Create(path)
		if err != nil {
			return err
		}

		// Encode blockImageMinimal as JPEG
		err = jpeg.Encode(fileWriter, q.blockImageMinimal, nil)
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

// decode reconstructs the quadtree structure from a zip file
func (q *QuadtreeElement) decode(path string, file *zip.File, remainingHeight int) error {
	// If path is empty a leaf has been reached
	if path == "" {
		// Read image from zipFile
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
		// Reconstruct blockImage by scaling fileImage up from BlockSize
		upsamplingInterpolator, err := getInterpolator(q.config.Quadtree.UpsamplingInterpolator)
		if err != nil {
			panic(err)
		}
		q.blockImage = utils.Scale(fileImageRGBA, q.baseImage.Bounds(), upsamplingInterpolator).(*image.RGBA)

		return nil
	}

	// Abort if the minimal tree height was reached and no leaf was detected yet
	if remainingHeight == 0 {
		return fmt.Errorf("further partitioning according to path %s would lead to remaining height being smaller than 0 in %s", path, q.id)
	}

	// If children haven't been created yet, create them
	if len(q.children) != ChildCount {
		// TODO: More or less duplicate code fragment
		for i := 0; i < ChildCount; i++ {
			var xStart, yStart, xEnd, yEnd int

			// Set x coordinates
			if i&1 == 0 {
				// Left block
				xStart = q.baseImage.Bounds().Min.X
				xEnd = q.baseImage.Bounds().Min.X + q.baseImage.Bounds().Dx()/2
			} else {
				// Right block
				xStart = q.baseImage.Bounds().Min.X + q.baseImage.Bounds().Dx()/2
				xEnd = q.baseImage.Bounds().Max.X
			}

			// Set y coordinates
			if i&2 == 0 {
				// Upper block
				yStart = q.baseImage.Bounds().Min.Y
				yEnd = q.baseImage.Bounds().Min.Y + q.baseImage.Bounds().Dy()/2
			} else {
				// Lower block
				yStart = q.baseImage.Bounds().Min.Y + q.baseImage.Bounds().Dy()/2
				yEnd = q.baseImage.Bounds().Max.Y
			}

			// Copy BaseImage section to sub image
			childImage := image.NewRGBA(image.Rect(xStart, yStart, xEnd, yEnd))

			// Create and append child without using NewQuadtreeElement as the block images are irrelevant during decoding
			child := &QuadtreeElement{
				id:        q.id + strconv.Itoa(i),
				baseImage: childImage,
				config:    q.config,
			}
			q.children = append(q.children, child)
		}
	}

	// Get next child from path
	splitPath := strings.Split(path, "/")
	childId, err := strconv.Atoi(splitPath[0])
	if err != nil {
		return err
	}

	// Sanity check childId
	if childId >= ChildCount {
		return fmt.Errorf("childId %d is greater than child count (%d)", childId, ChildCount)
	}

	// Recurse into next child
	recursePath := strings.Join(splitPath[1:], "/")
	return q.children[childId].decode(recursePath, file, remainingHeight-1)
}

// visualize returns its own blockImage if it has no children, else it returns its childrens blockImages
func (q *QuadtreeElement) visualize() []VisualizationElement {
	visualizations := make([]VisualizationElement, 0)

	if len(q.children) == 0 {
		visualizations = append(visualizations, VisualizationElement{image: q.blockImage, canBeSkipped: q.canBeSkipped})
	} else {
		for _, child := range q.children {
			visualizations = append(visualizations, child.visualize()...)
		}
	}

	return visualizations
}

// getInterpolator returns the correct interpolation algorithm for an interpolatorId from interpolators
func getInterpolator(interpolatorId string) (drawX.Interpolator, error) {
	interpolator, ok := interpolators[interpolatorId]
	var err error
	if !ok {
		err = fmt.Errorf("interpolator id not found: %q", interpolatorId)
	}
	return interpolator, err
}
