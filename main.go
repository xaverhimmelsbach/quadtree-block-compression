package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/xaverhimmelsbach/quadtree-block-compression/quadtreeImage"
)

// readImage Takes the path to an image file in the file system and returns the decoded image
func readImage(path string) (img image.Image, err error) {
	file, err := os.Open(path)
	if err != nil {
		return img, err
	}

	defer file.Close()
	img, _, err = image.Decode(file)
	return img, err
}

func main() {
	// Parse arguments
	inputPath := flag.String("input", "", "Image to encode as quadtree")
	outputPath := flag.String("output", "", "Path to write encoded file to")
	flag.Parse()

	// Read image from file system
	img, err := readImage(*inputPath)
	if err != nil {
		panic(err)
	}

	// Create quadtree image representation
	quadtreeImage := quadtreeImage.QuadtreeImage{
		BaseImage: img,
	}

	// Partition image into a quadtree structure
	quadtreeImage.Partition()

	// Encode quadtree structure
	quadtreeImage.Encode()

	// Visualize quadtree structure
	quadtreeImage.Visualize(*outputPath)

	// Write quadtree structure to output
	quadtreeImage.WriteFile(*outputPath)

	fmt.Printf("Encoded %q as a quadtree image and wrote it to %q", *inputPath, *outputPath)
}
