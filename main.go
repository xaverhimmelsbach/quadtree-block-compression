package main

import (
	"flag"
	"fmt"

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

	// Read image from file system
	img, err := utils.ReadImage(*inputPath)
	if err != nil {
		panic(err)
	}

	// Create quadtree image representation
	quadtreeRoot := quadtreeImage.NewQuadtreeImage(img)

	// Partition image into a quadtree structure
	quadtreeRoot.Partition()

	// Encode quadtree structure
	quadtreeRoot.Encode()

	// Visualize quadtree structure
	baseVisualization, paddedVisualization, baseBlockVisualization, paddedBlockVisualization, err := quadtreeRoot.Visualize(*outputPath)
	if err != nil {
		panic(err)
	}
	utils.WriteImage("visualizationBase.png", baseVisualization)
	utils.WriteImage("visualizationPadded.png", paddedVisualization)
	utils.WriteImage("visualizationBlockBase.png", baseBlockVisualization)
	utils.WriteImage("visualizationBlockPadded.png", paddedBlockVisualization)

	// Write quadtree structure to output
	quadtreeRoot.WriteFile(*outputPath)

	fmt.Printf("Encoded %q as a quadtree image and wrote it to %q", *inputPath, *outputPath)
}
