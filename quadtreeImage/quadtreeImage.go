package quadtreeImage

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"io/ioutil"
	"math"
	"strconv"
	"strings"

	"github.com/xaverhimmelsbach/quadtree-block-compression/config"
	"github.com/xaverhimmelsbach/quadtree-block-compression/utils"
)

// QuadtreeImage holds and manages a quadtree image
type QuadtreeImage struct {
	// Original image
	baseImage image.Image
	// Original image with added padding to make it quadratic
	paddedImage image.Image
	// Root node of the quadtree
	root *QuadtreeElement
	// Program configuration
	config *config.Config
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
	rootImage := image.NewRGBA(image.Rect(0, 0, q.paddedImage.Bounds().Max.X, q.paddedImage.Bounds().Max.Y))
	draw.Draw(rootImage, rootImage.Bounds(), q.paddedImage, q.paddedImage.Bounds().Min, draw.Src)
	q.root = NewQuadtreeElement("", rootImage, q.baseImage.Bounds(), q.config)

	// Start partitioning the quadtree
	q.root.partition()
}

// Encode encodes a quadtree image into a single buffer and returns it
func (q *QuadtreeImage) Encode() (io.Reader, error) {
	buffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buffer)

	// TODO: What happens if the first child can already encode the whole picture (e.g. solid color)?
	// Encode the tree root, which recurses further down the quadtree if needed
	err := q.root.encode(zipWriter)
	if err != nil {
		return buffer, err
	}

	fileWriter, err := zipWriter.Create(MetaFile)
	if err != nil {
		return buffer, err
	}

	treeHeight, err := q.getHeight()
	if err != nil {
		return buffer, err
	}

	width := q.baseImage.Bounds().Dx()
	height := q.baseImage.Bounds().Dy()

	// Write metadata
	fileWriter.Write([]byte(strconv.Itoa(treeHeight) + "\n" +
		strconv.Itoa(width) + "\n" +
		strconv.Itoa(height)))

	// Close zipWriter explicitly to flush all files to buffer
	zipWriter.Close()
	return buffer, err
}

// Decode decodes an encoded quadtree image and populates a quadtree with it
func Decode(quadtreePath string, outputPath string, cfg *config.Config) (*QuadtreeImage, error) {
	zipReader, err := zip.OpenReader(quadtreePath)
	if err != nil {
		return nil, err
	}

	metaFile, err := zipReader.Open(MetaFile)
	if err != nil {
		return nil, err
	}

	metaBytes, err := ioutil.ReadAll(metaFile)
	if err != nil {
		return nil, err
	}

	// Parse metadata
	meta := strings.Split(string(metaBytes), "\n")
	if len(meta) != 3 {
		return nil, fmt.Errorf("meta file contained %d newline-seperated values instead of three", len(meta))
	}

	treeHeight, err := strconv.Atoi(meta[0])
	if err != nil {
		return nil, err
	}

	width, err := strconv.Atoi(meta[1])
	if err != nil {
		return nil, err
	}

	height, err := strconv.Atoi(meta[2])
	if err != nil {
		return nil, err
	}

	baseImage := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create QuadtreeImage
	qti := NewQuadtreeImage(baseImage, cfg)

	// Create root manually to avoid calling its partition method
	qti.root = &QuadtreeElement{
		id:        "",
		config:    cfg,
		baseImage: qti.paddedImage,
	}

	// Iterate over zips contents and decode them
	for _, file := range zipReader.File {
		// Skip metadata
		if file.Name == MetaFile {
			continue
		}

		// Decode file into quadtree
		err = qti.root.decode(file.Name, file, treeHeight)
		if err != nil {
			return qti, err
		}
	}

	return qti, nil
}

// GetBlockImage creates a representation of the image encoded in the quadtree.
// If padded is true, the padding area around the original image is included as well.
func (q *QuadtreeImage) GetBlockImage(padded bool) image.Image {
	visualizations := q.root.visualize()

	// Choose correct inputImage
	var inputBounds image.Rectangle
	if padded {
		inputBounds = q.paddedImage.Bounds()
	} else {
		inputBounds = q.baseImage.Bounds()
	}

	// Setup bounds of blockImage
	blockImage := image.NewRGBA(image.Rect(0, 0, inputBounds.Dx(), inputBounds.Dy()))

	// Draw blocks of quadtree leaves onto blockimage
	for _, visualization := range visualizations {
		// Skip skippable blocks for unpadded images
		if padded || !visualization.canBeSkipped {
			draw.Draw(blockImage, visualization.image.Bounds(), visualization.image, visualization.image.Bounds().Min, draw.Src)
		}
	}

	return blockImage
}

// GetBoxImage creates a representation of the bounding boxes of the quadtree.
// If padded is true, the padding area around the original image is included as well.
func (q *QuadtreeImage) GetBoxImage(padded bool) image.Image {
	visualizations := q.root.visualize()

	// Choose correct inputImage
	var inputImage image.Image
	if padded {
		inputImage = q.paddedImage
	} else {
		inputImage = q.baseImage
	}

	// Copy inputImage onto boxImage
	boxImage := image.NewRGBA(image.Rect(0, 0, inputImage.Bounds().Dx(), inputImage.Bounds().Dy()))
	draw.Draw(boxImage, boxImage.Bounds(), inputImage, boxImage.Bounds().Min, draw.Src)

	// Draw bounding boxes
	for _, visualization := range visualizations {
		// Skip skippable boxes for unpadded images
		if padded || !visualization.canBeSkipped {
			utils.Rectangle(boxImage, visualization.image.Bounds().Min.X, visualization.image.Bounds().Max.X, visualization.image.Bounds().Min.Y, visualization.image.Bounds().Max.Y, color.RGBA{R: 255, A: 255})
		}
	}

	return boxImage
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

// getHeight returns how high the quadtree would need to be to have children of size BlockSize as leaves
func (q *QuadtreeImage) getHeight() (int, error) {
	// Ensure that paddedImage is quadratic
	dx := q.paddedImage.Bounds().Dx()
	dy := q.paddedImage.Bounds().Dy()

	if dx != dy {
		return 0, fmt.Errorf("padded image is not quadratic (width: %d, height: %d)", dx, dy)
	}

	// How many blocks would the tree be made up of in the worst case?
	blockCount := float64(dx) / float64(BlockSize)
	// How often would the tree need to partition to get to blocks of size BlockSize?
	return int(math.Log2(blockCount)), nil
}
