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
	configPath := flag.String("config", "", "Path to read program config from")
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
			outputFilename := "output" + path.Ext(*outputPath)

			inputFile, err := os.Open(*inputPath)
			if err != nil {
				panic(err)
			}

			(*analyticsFiles)[inputFilename] = inputFile
			(*analyticsFiles)[outputFilename] = &encodedClone
		}

		writeAnalytics(analyticsFiles, *analyticsDir, cfg.VisualizationConfig.Enable)

	case filetype.IsArchive(inputBuffer):
		fmt.Println("Decoding quadtree file")
		decoded, analyticsFiles, err := quadtreeImage.Decode(*inputPath, *outputPath, cfg)
		if err != nil {
			panic(err)
		}

		// Create clone of decoded to write it to analytics as well
		var decodedClone bytes.Buffer
		decodedTee := io.TeeReader(decoded, &decodedClone)

		// decodedTee has to be read before encodedClone
		err = utils.WriteFile(*outputPath, decodedTee)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Decoded %s and wrote it to %s", *inputPath, *outputPath)

		// Write input and output files to analytics
		if cfg.VisualizationConfig.Enable {
			inputFilename := "input" + path.Ext(*inputPath)
			outputFilename := "output" + path.Ext(*outputPath)

			inputFile, err := os.Open(*inputPath)
			if err != nil {
				panic(err)
			}

			(*analyticsFiles)[inputFilename] = inputFile
			(*analyticsFiles)[outputFilename] = &decodedClone
		}

		writeAnalytics(analyticsFiles, *analyticsDir, cfg.VisualizationConfig.Enable)
	default:
		panic("filetype is neither image nor archive")
	}
}

func directoryExists(directory string) (bool, error) {
	if _, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return true, err
		}
	}
	return true, nil
}

func writeAnalytics(analyticsFiles *map[string]io.Reader, analyticsDir string, analyticsEnabled bool) {
	if analyticsEnabled && len(analyticsDir) > 0 {
		// Create sub directory with current timestamp for currentAnalytics
		timestamp := fmt.Sprint(time.Now().Unix())
		currentAnalyticsDir := path.Join(analyticsDir, timestamp)

		// Try to create valid directory, if one already exists for the current timestamp by appending a number
		i := 0

		exists, err := directoryExists(currentAnalyticsDir)
		for exists {
			if err != nil {
				panic(err)
			}

			currentAnalyticsDir = path.Join(analyticsDir, fmt.Sprintf("%s_%d", timestamp, i))
			i = i + 1

			// A bit dumb to have this line twice, but err must be checked...
			exists, err = directoryExists(currentAnalyticsDir)
		}

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
}
