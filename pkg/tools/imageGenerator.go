package tools

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"math"
	"math/rand"
	"os"
	"os/signal"

	"github.com/xaverhimmelsbach/quadtree-block-compression/pkg/utils"
)

type Result[T any] struct {
	image io.Reader
	meta  T
}

type Generator[T any] interface {
	Evaluate(image.Image)
	GetResults() []Result[T]
}

type LargestBlockGenerator struct {
	currentLargestImage     io.Reader
	currentLargestImageSize int
}

type SmallestBlockGenerator struct {
	currentSmallestImage     io.Reader
	currentSmallestImageSize int
}

type FilesizeMeta struct {
	Size int
}

func NewLargestBlockGenerator() *LargestBlockGenerator {
	return new(LargestBlockGenerator)
}

func (g *LargestBlockGenerator) Evaluate(img image.Image) {
	buffer := new(bytes.Buffer)
	jpeg.Encode(buffer, img, nil)
	size := len(buffer.Bytes())
	if size > g.currentLargestImageSize {
		g.currentLargestImage = buffer
		g.currentLargestImageSize = size
	}
}

func (g *LargestBlockGenerator) GetResults() []Result[FilesizeMeta] {
	return []Result[FilesizeMeta]{
		{
			image: g.currentLargestImage,
			meta: FilesizeMeta{
				Size: g.currentLargestImageSize,
			},
		},
	}
}

func NewSmallestBlockGenerator() *SmallestBlockGenerator {
	generator := new(SmallestBlockGenerator)
	generator.currentSmallestImageSize = math.MaxInt
	return generator
}

func (g *SmallestBlockGenerator) Evaluate(img image.Image) {
	buffer := new(bytes.Buffer)
	jpeg.Encode(buffer, img, nil)
	size := len(buffer.Bytes())

	if size < g.currentSmallestImageSize {
		g.currentSmallestImage = buffer
		g.currentSmallestImageSize = size
	}
}

func (g *SmallestBlockGenerator) GetResults() []Result[FilesizeMeta] {
	return []Result[FilesizeMeta]{
		{
			image: g.currentSmallestImage,
			meta: FilesizeMeta{
				Size: g.currentSmallestImageSize,
			},
		},
	}
}

func generate() {
	filesizeGenerators := []Generator[FilesizeMeta]{
		NewLargestBlockGenerator(),
		NewSmallestBlockGenerator(),
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

loop:
	for /*samples := 0; samples < 100; samples++*/ {
		img := image.NewRGBA(image.Rect(0, 0, 8, 8))
		for x := 0; x < 8; x++ {
			for y := 0; y < 8; y++ {
				img.Set(x, y, color.RGBA{
					R: uint8(rand.Intn(256)),
					G: uint8(rand.Intn(256)),
					B: uint8(rand.Intn(256)),
					A: 255,
				})
			}
		}
		for _, g := range filesizeGenerators {
			g.Evaluate(img)
		}

		select {
		case <-c:
			break loop
		default:
		}
	}

	for i, g := range filesizeGenerators {
		results := g.GetResults()
		for j, r := range results {
			filename := fmt.Sprintf("./%d_%d_%d.jpg", i, j, r.meta.Size)
			if r.image != nil {
				utils.WriteFile(filename, r.image)
				fmt.Printf("Wrote file %s with filesize %d\n", filename, r.meta.Size)
			} else {
				fmt.Printf("%s is nil\n", filename)
			}
		}
	}
}

/* func main() {
	generate()
} */
