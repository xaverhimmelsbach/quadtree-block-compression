package main

import (
	"flag"
	"fmt"

	"github.com/xaverhimmelsbach/quadtree-block-compression/quadtreeImage"
	"github.com/xaverhimmelsbach/quadtree-block-compression/utils"
)

func main() {
	// Parse arguments
	inputPath := flag.String("input", "", "Image to encode as quadtree")
	outputPath := flag.String("output", "", "Path to write encoded file to")
	flag.Parse()

	// Read image from file system
	img, err := utils.ReadImage(*inputPath)
	if err != nil {
		panic(err)
	}

	// Create quadtree image representation
	quadtreeImage := quadtreeImage.QuadtreeImage{}

	// Partition image into a quadtree structure
	quadtreeImage.Partition(img)

	// Encode quadtree structure
	quadtreeImage.Encode()

	// Visualize quadtree structure
	baseVisualization, paddedVisualization, baseBlockVisualization, paddedBlockVisualization, err := quadtreeImage.Visualize(*outputPath)
	if err != nil {
		panic(err)
	}
	utils.WriteImage("visualizationBase.png", baseVisualization)
	utils.WriteImage("visualizationPadded.png", paddedVisualization)
	utils.WriteImage("visualizationBlockBase.png", baseBlockVisualization)
	utils.WriteImage("visualizationBlockPadded.png", paddedBlockVisualization)

	// Write quadtree structure to output
	quadtreeImage.WriteFile(*outputPath)

	fmt.Printf("Encoded %q as a quadtree image and wrote it to %q", *inputPath, *outputPath)
}
