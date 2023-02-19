package quadtreeImage

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/xaverhimmelsbach/quadtree-block-compression/pkg/config"
	"github.com/xaverhimmelsbach/quadtree-block-compression/pkg/utils"
)

// QuadtreeImage holds and manages a quadtree image
type QuadtreeImage struct {
	// Original image
	baseImage image.Image
	// Original image with added padding to make it quadratic
	paddedImage image.Image
	// Root node of the quadtree
	root *QuadtreeElement
	// List of all currently existing quadtree blocks of size BlockSize
	existingBlocks **[]*image.Image
	// Program configuration
	config *config.Config
}

// NewQuadtreeImage constructs a well-formed instance of QuadtreeImage from a baseImage
func NewQuadtreeImage(baseImage image.Image, cfg *config.Config) *QuadtreeImage {
	qti := new(QuadtreeImage)

	qti.config = cfg
	qti.baseImage = baseImage
	qti.paddedImage = qti.pad()

	// Create pointer to shared block list
	blocks := new([]*image.Image)
	// Create pointer to pointer so that quadtreeElements can modify the shared list
	qti.existingBlocks = &blocks

	return qti
}

// Partition splits the BaseImage into an appropriate number of sub images and calls their partition method
// TODO: Make this private and call it from Encode. Also rework Encode to work as a static function and handle creating the quadtree in there.
func (q *QuadtreeImage) Partition() {
	// Create root of the quadtree
	rootImage := image.NewRGBA(image.Rect(0, 0, q.paddedImage.Bounds().Max.X, q.paddedImage.Bounds().Max.Y))
	draw.Draw(rootImage, rootImage.Bounds(), q.paddedImage, q.paddedImage.Bounds().Min, draw.Src)
	globalBounds := q.baseImage.Bounds()
	q.root = NewQuadtreeElement("", rootImage, &globalBounds, q.existingBlocks, q.config)

	// WaitGroup for use in parallelized partition
	var wg sync.WaitGroup

	if q.config.Encoding.Parallelism {
		wg.Add(1)
	}

	// Start partitioning the quadtree
	q.root.partition(&wg)

	wg.Wait()
}

// Encode encodes a quadtree image into a single buffer and returns it
func (q *QuadtreeImage) Encode(archiveMode ArchiveMode) (io.Reader, *map[string]io.Reader, error) {
	fileBuffer := new(bytes.Buffer)
	analyticsFiles := make(map[string]io.Reader)

	// TODO: Do this right after partitioning
	if q.config.VisualizationConfig.Enable {
		boxVisualization := q.GetBoxImage(false)
		boxVisualizationPadded := q.GetBoxImage(true)
		blockVisualization := q.GetBlockImage(false)
		blockVisualizationPadded := q.GetBlockImage(true)

		boxVisualizationBuffer := new(bytes.Buffer)
		utils.WriteImage(boxVisualization, boxVisualizationBuffer, ".jpg")
		boxVisualizationPaddedBuffer := new(bytes.Buffer)
		utils.WriteImage(boxVisualizationPadded, boxVisualizationPaddedBuffer, ".jpg")
		blockVisualizationBuffer := new(bytes.Buffer)
		utils.WriteImage(blockVisualization, blockVisualizationBuffer, ".jpg")
		blockVisualizationPaddedBuffer := new(bytes.Buffer)
		utils.WriteImage(blockVisualizationPadded, blockVisualizationPaddedBuffer, ".jpg")

		analyticsFiles["encodedBoxVisualization.jpg"] = boxVisualizationBuffer
		analyticsFiles["encodedBoxVisualizationPadded.jpg"] = boxVisualizationPaddedBuffer
		analyticsFiles["encodedBlockVisualization.jpg"] = blockVisualizationBuffer
		analyticsFiles["encodedBlockVisualizationPadded.jpg"] = blockVisualizationPaddedBuffer
	}

	archiveWriter, err := NewArchiveWriter(archiveMode, fileBuffer)
	if err != nil {
		return fileBuffer, &analyticsFiles, err
	}

	// Keep map of encoded blocks and their path in the archive for deduplication
	encodedBlockPaths := make(map[*image.Image]string)

	// TODO: What happens if the first child can already encode the whole picture (e.g. solid color)?
	// Encode the tree root, which recurses further down the quadtree if needed
	err = q.root.encode(archiveWriter, &encodedBlockPaths)
	if err != nil {
		return fileBuffer, &analyticsFiles, err
	}

	treeHeight, err := q.getHeight()
	if err != nil {
		return fileBuffer, &analyticsFiles, err
	}

	width := q.baseImage.Bounds().Dx()
	height := q.baseImage.Bounds().Dy()

	// Write metadata
	metaBuffer := new(bytes.Buffer)
	metaBuffer.Write([]byte(strconv.Itoa(treeHeight) + "\n" +
		strconv.Itoa(width) + "\n" +
		strconv.Itoa(height)))

	err = archiveWriter.WriteFile(MetaFile, metaBuffer)
	if err != nil {
		return fileBuffer, &analyticsFiles, err
	}

	// Close archiveWriter explicitly to flush all files to buffer
	err = archiveWriter.Close()
	return fileBuffer, &analyticsFiles, err
}

// Decode decodes an encoded quadtree image and populates a quadtree with it
func Decode(quadtreePath string, outputPath string, cfg *config.Config) (io.Reader, *map[string]io.Reader, error) {
	analyticsFiles := make(map[string]io.Reader)

	archiveReader, err := OpenArchiveReader(quadtreePath)
	if err != nil {
		return nil, &analyticsFiles, err
	}

	// Parse metadata
	metaBytes, err := archiveReader.Open(MetaFile)
	if err != nil {
		return nil, &analyticsFiles, err
	}

	meta := strings.Split(string(*metaBytes), "\n")
	if len(meta) != 3 {
		return nil, &analyticsFiles, fmt.Errorf("meta file contained %d newline-seperated values instead of three", len(meta))
	}

	treeHeight, err := strconv.Atoi(meta[0])
	if err != nil {
		return nil, &analyticsFiles, err
	}

	width, err := strconv.Atoi(meta[1])
	if err != nil {
		return nil, &analyticsFiles, err
	}

	height, err := strconv.Atoi(meta[2])
	if err != nil {
		return nil, &analyticsFiles, err
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

	// Iterate over archive contents and decode them
	for filename, fileContents := range archiveReader.Files() {
		// Skip metadata file
		if filename == MetaFile {
			continue
		}

		// Decode file into quadtree
		err = qti.root.decode(filename, fileContents, treeHeight, archiveReader)
		if err != nil {
			return nil, &analyticsFiles, err
		}
	}

	if qti.config.VisualizationConfig.Enable {
		boxVisualization := qti.GetBoxImage(false)
		boxVisualizationPadded := qti.GetBoxImage(true)
		blockVisualization := qti.GetBlockImage(false)
		blockVisualizationPadded := qti.GetBlockImage(true)

		boxVisualizationBuffer := new(bytes.Buffer)
		utils.WriteImage(boxVisualization, boxVisualizationBuffer, ".jpg")
		boxVisualizationPaddedBuffer := new(bytes.Buffer)
		utils.WriteImage(boxVisualizationPadded, boxVisualizationPaddedBuffer, ".jpg")
		blockVisualizationBuffer := new(bytes.Buffer)
		utils.WriteImage(blockVisualization, blockVisualizationBuffer, ".jpg")
		blockVisualizationPaddedBuffer := new(bytes.Buffer)
		utils.WriteImage(blockVisualizationPadded, blockVisualizationPaddedBuffer, ".jpg")

		analyticsFiles["decodedBoxVisualization.jpg"] = boxVisualizationBuffer
		analyticsFiles["decodedBoxVisualizationPadded.jpg"] = boxVisualizationPaddedBuffer
		analyticsFiles["decodedBlockVisualization.jpg"] = blockVisualizationBuffer
		analyticsFiles["decodedBlockVisualizationPadded.jpg"] = blockVisualizationPaddedBuffer
	}

	fileBuffer := new(bytes.Buffer)
	utils.WriteImage(qti.GetBlockImage(false), fileBuffer, ".jpg")

	return fileBuffer, &analyticsFiles, nil
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
		if visualization.image != nil && (padded || !visualization.canBeSkipped) {
			draw.Draw(blockImage, visualization.image.Bounds(), visualization.image, visualization.image.Bounds().Min, draw.Src)
		}
	}

	return blockImage
}

// GetBoxImage creates a representation of the bounding boxes of the quadtree.
// If padded is true, the padding area around the original image is included as well.
func (q *QuadtreeImage) GetBoxImage(padded bool) image.Image {
	visualizations := q.root.visualize()

	// Get background image
	inputImage := q.GetBlockImage(padded)

	// Copy inputImage onto boxImage
	boxImage := image.NewRGBA(image.Rect(0, 0, inputImage.Bounds().Dx(), inputImage.Bounds().Dy()))
	draw.Draw(boxImage, boxImage.Bounds(), inputImage, boxImage.Bounds().Min, draw.Src)

	// Draw bounding boxes
	for _, visualization := range visualizations {
		// Skip skippable boxes for unpadded images
		if visualization.image != nil && (padded || !visualization.canBeSkipped) {
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
