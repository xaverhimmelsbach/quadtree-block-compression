package quadtreeImage

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"strconv"

	"github.com/xaverhimmelsbach/quadtree-block-compression/config"
	"github.com/xaverhimmelsbach/quadtree-block-compression/utils"
)

type QuadtreeImage struct {
	baseImage   image.Image
	paddedImage image.Image
	child       *QuadtreeElement
	config      *config.Config
}

// NewQuadtreeImage constructs a well-formed instance of QuadtreeImage from a baseImage
func NewQuadtreeImage(baseImage image.Image, cfg *config.Config) *QuadtreeImage {
	qti := new(QuadtreeImage)

	qti.config = cfg
	qti.baseImage = baseImage
	qti.paddedImage = qti.pad()

	return qti
}

// Partition splits the BaseImage into an appropriate number of sub images and calls their partition method
func (q *QuadtreeImage) Partition() {
	// Create root of the quadtree
	childImage := image.NewRGBA(image.Rect(0, 0, q.paddedImage.Bounds().Max.X, q.paddedImage.Bounds().Max.Y))
	draw.Draw(childImage, childImage.Bounds(), q.paddedImage, q.paddedImage.Bounds().Min, draw.Src)
	q.child = NewQuadtreeElement("", childImage, q.baseImage.Bounds(), q.config)

	// Start partitioning the quadtree
	q.child.partition()
}

// Encode encodes a quadtree image into single file and writes it to the file system
func (q *QuadtreeImage) Encode(filePath string) error {
	targetFile, err := os.Create(filePath)
	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(targetFile)
	defer zipWriter.Close()

	err = q.child.encode(zipWriter)
	if err != nil {
		return err
	}

	fileWriter, err := zipWriter.Create(MetaFile)
	if err != nil {
		return err
	}

	treeHeight, err := q.getHeight()
	if err != nil {
		return err
	}

	width := q.baseImage.Bounds().Dx()
	height := q.baseImage.Bounds().Dy()

	fileWriter.Write([]byte(strconv.Itoa(treeHeight) + "\n" +
		strconv.Itoa(width) + "\n" +
		strconv.Itoa(height)))

	return err
}

// Visualize draws the bounding boxes of all Children onto a copy of the BaseImage and of the PaddedImage.
// It also draws the upsampled JPEG blocks to show how the encoded result would look
func (q *QuadtreeImage) Visualize() (image.Image, image.Image, image.Image, image.Image, error) {
	images := q.child.visualize()
	baseBounds := q.baseImage.Bounds()
	paddedBounds := q.paddedImage.Bounds()

	baseImage := image.NewRGBA(image.Rect(0, 0, baseBounds.Dx(), baseBounds.Dy()))
	draw.Draw(baseImage, baseImage.Bounds(), q.baseImage, baseBounds.Min, draw.Src)

	paddedImage := image.NewRGBA(image.Rect(0, 0, paddedBounds.Dx(), paddedBounds.Dy()))
	draw.Draw(paddedImage, paddedImage.Bounds(), q.paddedImage, paddedBounds.Min, draw.Src)

	baseImageBlocks := image.NewRGBA(image.Rect(0, 0, baseBounds.Dx(), baseBounds.Dy()))

	paddedImageBlocks := image.NewRGBA(image.Rect(0, 0, paddedBounds.Dx(), paddedBounds.Dy()))

	for _, img := range images {
		// Draw bounding boxes
		utils.Rectangle(baseImage, img.Bounds().Min.X, img.Bounds().Max.X, img.Bounds().Min.Y, img.Bounds().Max.Y, color.RGBA{R: 255, A: 255})
		utils.Rectangle(paddedImage, img.Bounds().Min.X, img.Bounds().Max.X, img.Bounds().Min.Y, img.Bounds().Max.Y, color.RGBA{R: 255, A: 255})

		// Combine separate upscaled blocks into whole images
		draw.Draw(baseImageBlocks, img.Bounds(), img, img.Bounds().Min, draw.Src)
		draw.Draw(paddedImageBlocks, img.Bounds(), img, img.Bounds().Min, draw.Src)
	}

	return baseImage, paddedImage, baseImageBlocks, paddedImageBlocks, nil
}

// pad adds transparent padding to a copy of BaseImage to make it a square with an edge length that can be divided by a multiple of four to get a JPEG block
func (q *QuadtreeImage) pad() image.Image {
	baseBounds := q.baseImage.Bounds()
	var longerSideLength int
	paddedSideLength := BlockSize

	// Find the longer side of X and Y
	if baseBounds.Dx() > baseBounds.Dy() {
		longerSideLength = baseBounds.Dx()
	} else {
		longerSideLength = baseBounds.Dy()
	}

	// Pad until the padding is greater than both sides of the BaseImage
	for paddedSideLength < longerSideLength {
		paddedSideLength *= ChildCount
	}

	// Copy BaseImage over padded image
	paddedImage := image.NewRGBA(image.Rect(0, 0, paddedSideLength, paddedSideLength))
	draw.Draw(paddedImage, paddedImage.Bounds(), q.baseImage, q.baseImage.Bounds().Min, draw.Src)

	utils.FillSpace(paddedImage, q.baseImage.Bounds())

	return paddedImage
}

// getHeight returns how high the quadtree would need to be, to have children of size BlockSize as leaves
func (q *QuadtreeImage) getHeight() (int, error) {
	// Make sure that paddedImage is quadratic
	dx := q.paddedImage.Bounds().Dx()
	dy := q.paddedImage.Bounds().Dy()

	if dx != dy {
		return 0, fmt.Errorf("padded image is not quadratic (width: %d, height: %d)", dx, dy)
	}

	// How many blocks would the tree be made up of in the worst case?
	blockCount := float64(dx) / float64(BlockSize)
	// How often the tree would need to partition to get to blocks of size BlockSize?
	return int(math.Log2(blockCount)), nil
}
