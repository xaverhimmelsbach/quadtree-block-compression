package tools

import (
	"fmt"
	"image"
	"image/draw"

	drawX "golang.org/x/image/draw"

	"github.com/xaverhimmelsbach/quadtree-block-compression/utils"
)

// scaleTest scales an input image down and up repeatedly with different interpolators and prints the similarity of each combination
func scaleTest(img image.Image) {
	interpolators := map[string]drawX.Interpolator{
		"NearestNeighbor": drawX.NearestNeighbor,
		"ApproxBiLinear":  drawX.ApproxBiLinear,
		"BiLinear":        drawX.BiLinear,
		"CatmullRom":      drawX.CatmullRom,
	}

	imgRGBA := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X, img.Bounds().Max.Y))
	draw.Draw(imgRGBA, imgRGBA.Bounds(), img, img.Bounds().Min, draw.Src)

	for name, interpolator := range interpolators {
		for name2, interpolator2 := range interpolators {
			imgDownscaled := utils.Scale(imgRGBA, image.Rect(0, 0, 8, 8), interpolator)

			imgUpscaled := utils.Scale(imgDownscaled.(*image.RGBA), img.Bounds(), interpolator2)
			imgUpscaledRGBA := imgUpscaled.(*image.RGBA)

			// utils.WriteImage("downscaled.png", imgDownscaled)
			// utils.WriteImage("upscaled.png", imgUpscaled)

			similarity, err := utils.ComparePixelsWeighted(imgRGBA, imgUpscaledRGBA, imgRGBA.Bounds())
			if err != nil {
				panic(err)
			}

			fmt.Printf("%q -> %q: %f\n", name, name2, similarity)
		}
	}
}
