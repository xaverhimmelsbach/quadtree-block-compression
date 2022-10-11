package main

import (
	"flag"
	"fmt"
	"image"

	"github.com/xaverhimmelsbach/quadtree-block-compression/quadtreeImage"
)

// TODO: Implement
func readImage(path string) (image image.Image, err error) {
	return image, err
}

func main() {
	// Parse arguments
	inputPath := flag.String("input", "", "Image to encode as quadtree")
	outputPath := flag.String("output", "", "Path to write encoded file to")
	flag.Parse()

	// Read image from file system
	image, err := readImage(*inputPath)
	if err != nil {
		panic(err)
	}

	// Create quadtree image representation
	quadtreeImage := quadtreeImage.QuadtreeImage{
		BaseImage: image,
	}

	// Partition image into a quadtree structure
	quadtreeImage.Partition()

	// Encode quadtree structure
	quadtreeImage.Encode()

	// Write quadtree structure to output
	quadtreeImage.WriteFile(*outputPath)

	fmt.Printf("Encoded %q as a quadtree image and wrote it to %q", *inputPath, *outputPath)
}
