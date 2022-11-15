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
		fmt.Printf("Encoded %s as a quadtree image and wrote it to %s", *inputPath, *outputPath)

		// Visualize quadtree structure
		if cfg.VisualizationConfig.Enable {
			boxVisualization := quadtreeRoot.GetBoxImage(false)
			boxVisualizationPadded := quadtreeRoot.GetBoxImage(true)
			blockVisualization := quadtreeRoot.GetBlockImage(false)
			blockVisualizationPadded := quadtreeRoot.GetBlockImage(true)

			utils.WriteImage("boxVisualization.png", boxVisualization)
			utils.WriteImage("boxVisualizationPadded.png", boxVisualizationPadded)
			utils.WriteImage("blockVisualization.png", blockVisualization)
			utils.WriteImage("blockVisualizationPadded.png", blockVisualizationPadded)
		}
	case filetype.IsArchive(inputBuffer):
		fmt.Println("Decoding quadtree file")
		quadtreeRoot, err := quadtreeImage.Decode(*inputPath, *outputPath, cfg)
		if err != nil {
			panic(err)
		}

		decodedImage := quadtreeRoot.GetBlockImage(false)
		utils.WriteImage(*outputPath, decodedImage)
		fmt.Printf("Decoded %s and wrote it to %s", *inputPath, *outputPath)
	default:
		panic("filetype is neither image nor archive")
	}
}
