package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/h2non/filetype"
	"github.com/xaverhimmelsbach/quadtree-block-compression/config"
	"github.com/xaverhimmelsbach/quadtree-block-compression/quadtreeImage"
	"github.com/xaverhimmelsbach/quadtree-block-compression/utils"
)

func main() {
	// Parse arguments
	inputPath := flag.String("input", "", "Image to encode as quadtree")
	outputPath := flag.String("output", "", "Path to write encoded file to")
	configPath := flag.String("config", "config.yml", "Path to read program config from")
	flag.Parse()

	// Load config
	cfg, err := config.NewConfigFromFile(*configPath)
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg)

	// TODO: Reuse buffer for image reading
	inputBuffer, err := ioutil.ReadFile(*inputPath)
	if err != nil {
		panic(err)
	}

	switch true {
	case filetype.IsImage(inputBuffer):
		fmt.Println("Encoding image file")

		// Read image from file system
		img, err := utils.ReadImage(*inputPath)
		if err != nil {
			panic(err)
		}

		// Create quadtree image representation
		quadtreeRoot := quadtreeImage.NewQuadtreeImage(img, cfg)

		// Partition image into a quadtree structure
		quadtreeRoot.Partition()

		// Encode quadtree structure
		quadtreeRoot.Encode(*outputPath)

		// Visualize quadtree structure
		if cfg.VisualizationConfig.Enable {
			baseVisualization,
				paddedVisualization,
				baseBlockVisualization,
				paddedBlockVisualization,
				err := quadtreeRoot.Visualize()
			if err != nil {
				panic(err)
			}
			utils.WriteImage("visualizationBase.png", baseVisualization)
			utils.WriteImage("visualizationPadded.png", paddedVisualization)
			utils.WriteImage("visualizationBlockBase.png", baseBlockVisualization)
			utils.WriteImage("visualizationBlockPadded.png", paddedBlockVisualization)
		}

		// TODO: Write quadtree structure to output
		fmt.Printf("Encoded %q as a quadtree image and wrote it to %q", *inputPath, *outputPath)
	case filetype.IsArchive(inputBuffer):
		fmt.Println("Decoding fractal file")
		quadtreeRoot, err := quadtreeImage.Decode(*inputPath, *outputPath, cfg)
		if err != nil {
			panic(err)
		}

		decodedImage := quadtreeRoot.GetDecodedImage()
		utils.WriteImage(*outputPath, decodedImage)
	default:
		panic("filetype is neither image nor archive")
	}
}
