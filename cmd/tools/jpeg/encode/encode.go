package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"time"

	"github.com/xaverhimmelsbach/quadtree-block-compression/pkg/utils"
)

type EncodingResult struct {
	// Time in nanoseconds
	Time time.Duration `json:"time"`
	// Filesize in bytes
	Size int `json:"size"`
	// Potential errors
	Error string `json:"error"`
}

func main() {
	inputPath := flag.String("input", "", "Path to read image from")
	// outputPath := flag.String("output", "", "Path to write jpeg image to")
	flag.Parse()

	// Always print encoding result
	result := new(EncodingResult)
	defer printResult(result)

	img, err := utils.ReadImage(*inputPath)
	if err != nil {
		result.Error = err.Error()
		return
	}

	outputBuffer := new(bytes.Buffer)

	// Measure encoding time
	start := time.Now()
	// No jpeg options for now
	err = jpeg.Encode(outputBuffer, img, nil)
	elapsed := time.Since(start)
	result.Time = elapsed

	if err != nil {
		result.Error = err.Error()
		return
	}

	// Measure filesize
	outputBytes, err := ioutil.ReadAll(outputBuffer)
	result.Size = len(outputBytes)
	if err != nil {
		result.Error = err.Error()
		return
	}
}

func printResult(result *EncodingResult) {
	b, err := json.Marshal(result)
	output := string(b)
	// Fallback if marshalling fails
	if err != nil {
		output = fmt.Sprintf("{\"error\":%q}", err)
	}
	fmt.Println(output)
}
