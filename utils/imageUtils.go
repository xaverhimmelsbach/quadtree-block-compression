package utils

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// ReadImage Takes the path to an image file in the file system and returns the decoded image
func ReadImage(path string) (img image.Image, err error) {
	file, err := os.Open(path)
	if err != nil {
		return img, err
	}

	defer file.Close()
	img, _, err = image.Decode(file)
	return img, err
}
