package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/jpeg"
	"os"
	"time"
)

type DecodingResult struct {
	// Time in nanoseconds
	Time time.Duration `json:"time"`
	// Potential errors
	Error string `json:"error"`
}

func main() {
	inputPath := flag.String("input", "", "Path to read image from")
	// outputPath := flag.String("output", "", "Path to write jpeg image to")
	flag.Parse()

	// Always print decoding result
	result := new(DecodingResult)
	defer printResult(result)

	file, err := os.Open(*inputPath)
	if err != nil {
		result.Error = err.Error()
		return
	}

	// Measure decoding time
	start := time.Now()
	_, err = jpeg.Decode(file)
	elapsed := time.Since(start)
	result.Time = elapsed

	if err != nil {
		result.Error = err.Error()
		return
	}
}

func printResult(result *DecodingResult) {
	b, err := json.Marshal(result)
	output := string(b)
	// Fallback if marshalling fails
	if err != nil {
		output = fmt.Sprintf("{\"error\":%q}", err)
	}
	fmt.Println(output)
}
