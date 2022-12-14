package utils

import (
	"fmt"
	"image"
)

// ComparePixelsExact naively compares two images by checking how many of the pixels between them are identical.
// It returns a float that ranges between 0 (no matches) and 1 (identical pictures)
func ComparePixelsExact(imageA *image.RGBA, imageB *image.RGBA, globalBounds image.Rectangle) (float64, error) {
	// Ensure that images dimensions and origin points are equal
	if imageA.Bounds().Min.X != imageB.Bounds().Min.X ||
		imageA.Bounds().Min.Y != imageB.Bounds().Min.Y ||
		imageA.Bounds().Max.X != imageB.Bounds().Max.X ||
		imageA.Bounds().Max.Y != imageB.Bounds().Max.Y {
		return 0, fmt.Errorf("bounds for image A (%v) and image B (%v) do not match", imageA.Bounds(), imageB.Bounds())
	}

	matches := 0
	skipped := 0

	// Compare every pixel across both images
	for x := imageA.Bounds().Min.X; x < imageA.Bounds().Max.X; x++ {
		for y := imageA.Bounds().Min.Y; y < imageA.Bounds().Max.Y; y++ {

			// The padding is of no interest
			// TODO: Use function to calculate overlapping area instead of checking every pixel seperately
			if PointCollides(image.Point{X: x, Y: y}, globalBounds) {
				skipped++
				continue
			}

			aR, aG, aB, _ := imageA.At(x, y).RGBA()
			bR, bG, bB, _ := imageB.At(x, y).RGBA()

			if aR == bR && aG == bG && aB == bB {
				matches++
			}
		}
	}

	// If none of the checked pixels were inside bounds, signal that no further partitioning is required
	relevantPixels := imageA.Bounds().Dx()*imageA.Bounds().Dy() - skipped
	if relevantPixels <= 0 {
		return 1, nil
	}

	// Else compute similarity without OOB pixels
	similarity := float64(matches) / float64(relevantPixels)
	return similarity, nil
}

func ComparePixelsWeighted(imageA *image.RGBA, imageB *image.RGBA, globalBounds image.Rectangle) (float64, error) {
	// Ensure that images dimensions and origin points are equal
	if imageA.Bounds().Min.X != imageB.Bounds().Min.X ||
		imageA.Bounds().Min.Y != imageB.Bounds().Min.Y ||
		imageA.Bounds().Max.X != imageB.Bounds().Max.X ||
		imageA.Bounds().Max.Y != imageB.Bounds().Max.Y {
		return 0, fmt.Errorf("bounds for image A (%v) and image B (%v) do not match", imageA.Bounds(), imageB.Bounds())
	}

	var matches float64
	var skipped int

	// Compare every pixel across both images
	for x := imageA.Bounds().Min.X; x < imageA.Bounds().Max.X; x++ {
		for y := imageA.Bounds().Min.Y; y < imageA.Bounds().Max.Y; y++ {

			// The padding is of no interest
			if PointCollides(image.Point{X: x, Y: y}, globalBounds) {
				skipped++
				continue
			}

			aR, aG, aB, _ := imageA.At(x, y).RGBA()
			bR, bG, bB, _ := imageB.At(x, y).RGBA()

			// Use individual color channel weights
			if InRange(float64(aR), float64(bR), 1000*weightedRed) {
				matches += weightedRed
			}

			if InRange(float64(aG), float64(bG), 1000*weightedGreen) {
				matches += weightedGreen
			}

			if InRange(float64(aB), float64(bB), 1000*weightedBlue) {
				matches += weightedBlue
			}
		}
	}

	// If none of the checked pixels were inside bounds signal that no further partitioning is required
	// TODO: this shouldn't happen for quadtreeElements as they check for collisions before calling this function
	relevantPixels := imageA.Bounds().Dx()*imageA.Bounds().Dy() - skipped
	if relevantPixels <= 0 {
		fmt.Println("0 relevant pixels found. This shouldn't happen when calling ComparePixelsWeighted from compareImages on a quadtreeElement")
		return 1, nil
	}

	// Else compute similarity without OOB pixels
	similarity := matches / float64(relevantPixels)
	return similarity, nil
}

// PointCollides returns true if p collides with r
func PointCollides(p image.Point, r image.Rectangle) bool {
	return p.X < r.Min.X ||
		p.X > r.Max.X ||
		p.Y < r.Min.Y ||
		p.Y > r.Max.Y
}

// RectanglesCollide returns true if r1 and r2 collide
func RectanglesCollide(r1 image.Rectangle, r2 image.Rectangle) bool {
	return r1.Min.X < r2.Max.X &&
		r1.Max.X > r2.Min.X &&
		r1.Min.Y < r2.Max.Y &&
		r1.Max.Y > r2.Min.Y
}
