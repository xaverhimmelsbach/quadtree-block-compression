package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/h2non/filetype"
	"github.com/xaverhimmelsbach/quadtree-block-compression/pkg/config"
	"github.com/xaverhimmelsbach/quadtree-block-compression/pkg/quadtreeImage"
	"github.com/xaverhimmelsbach/quadtree-block-compression/pkg/utils"
)

func main() {
	// Parse arguments
	inputPath := flag.String("input", "", "Image to encode as quadtree")
	outputPath := flag.String("output", "", "Path to write encoded file to")
	configPath := flag.String("config", "config.yml", "Path to read program config from")
	analyticsDir := flag.String("analyticsDir", "", "Directory to write analytics to")
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
		encoded, analyticsFiles, err := quadtreeRoot.Encode(quadtreeImage.ArchiveMode(cfg.Encoding.ArchiveFormat))
		if err != nil {
			panic(err)
		}

		err = utils.WriteFile(*outputPath, encoded)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Encoded %s as a quadtree image and wrote it to %s\n", *inputPath, *outputPath)

		if cfg.VisualizationConfig.Enable && len(*analyticsDir) > 0 {
			// Create sub directory with current timestamp for currentAnalytics
			timestamp := fmt.Sprint(time.Now().Unix())
			currentAnalyticsDir := path.Join(*analyticsDir, timestamp)

			err = os.MkdirAll(currentAnalyticsDir, 0755)
			if err != nil {
				panic(err)
			}

			// TODO: Write input & output files to analyticsDir

			// Write encoding analytics if appropriate
			if len(*analyticsFiles) > 0 {
				for filename, reader := range *analyticsFiles {
					filepath := path.Join(currentAnalyticsDir, filename)
					err = utils.WriteFile(filepath, reader)
					if err != nil {
						panic(err)
					}
				}

				fmt.Printf("Wrote analytics files to %s\n", currentAnalyticsDir)
			}
		}

	case filetype.IsArchive(inputBuffer):
		fmt.Println("Decoding quadtree file")
		decodedImage, analyticsFiles, err := quadtreeImage.Decode(*inputPath, *outputPath, cfg)
		if err != nil {
			panic(err)
		}

		// TODO: Create function for writing analytics
		if cfg.VisualizationConfig.Enable && len(*analyticsDir) > 0 {
			// Create sub directory with current timestamp for currentAnalytics
			timestamp := fmt.Sprint(time.Now().Unix())
			currentAnalyticsDir := path.Join(*analyticsDir, timestamp)

			err = os.MkdirAll(currentAnalyticsDir, 0755)
			if err != nil {
				panic(err)
			}

			// TODO: Write input & output files to analyticsDir

			// Write encoding analytics if appropriate
			if len(*analyticsFiles) > 0 {
				for filename, reader := range *analyticsFiles {
					filepath := path.Join(currentAnalyticsDir, filename)
					err = utils.WriteFile(filepath, reader)
					if err != nil {
						panic(err)
					}
				}

				fmt.Printf("Wrote analytics files to %s\n", currentAnalyticsDir)
			}
		}

		utils.WriteImageToFile(decodedImage, *outputPath)
		fmt.Printf("Decoded %s and wrote it to %s", *inputPath, *outputPath)
	default:
		panic("filetype is neither image nor archive")
	}
}
