package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
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

		// Create clone of encoded to write it to analytics as well
		var encodedClone bytes.Buffer
		encodedTee := io.TeeReader(encoded, &encodedClone)

		// encodedTee has to be read before encodedClone
		err = utils.WriteFile(*outputPath, encodedTee)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Encoded %s as a quadtree image and wrote it to %s\n", *inputPath, *outputPath)

		// Write input and output files to analytics
		if cfg.VisualizationConfig.Enable {
			inputFilename := "input" + path.Ext(*inputPath)
			inputFile, err := os.Open(*inputPath)
			if err != nil {
				panic(err)
			}

			outputFilename := "output" + path.Ext(*outputPath)

			(*analyticsFiles)[inputFilename] = inputFile
			(*analyticsFiles)[outputFilename] = &encodedClone
		}

		writeAnalytics(analyticsFiles, *analyticsDir, cfg.VisualizationConfig.Enable)

	case filetype.IsArchive(inputBuffer):
		fmt.Println("Decoding quadtree file")
		decodedImage, analyticsFiles, err := quadtreeImage.Decode(*inputPath, *outputPath, cfg)
		if err != nil {
			panic(err)
		}

		utils.WriteImageToFile(decodedImage, *outputPath)
		fmt.Printf("Decoded %s and wrote it to %s", *inputPath, *outputPath)

		writeAnalytics(analyticsFiles, *analyticsDir, cfg.VisualizationConfig.Enable)
	default:
		panic("filetype is neither image nor archive")
	}
}

func writeAnalytics(analyticsFiles *map[string]io.Reader, analyticsDir string, analyticsEnabled bool) {
	if analyticsEnabled && len(analyticsDir) > 0 {
		// Create sub directory with current timestamp for currentAnalytics
		timestamp := fmt.Sprint(time.Now().Unix())
		currentAnalyticsDir := path.Join(analyticsDir, timestamp)

		err := os.MkdirAll(currentAnalyticsDir, 0755)
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
}
