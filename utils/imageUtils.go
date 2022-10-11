package utils

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"
	"path"
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

// WriteImage encodes an image as either PNG or JPEG according to the file extension and writes it to the provided path
func WriteImage(filePath string, img image.Image) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	extension := path.Ext(filePath)
	switch extension {
	case ".jpg":
		fallthrough
	case ".jpeg":
		err = jpeg.Encode(file, img, nil)
	case ".png":
		err = png.Encode(file, img)
	default:
		err = fmt.Errorf("unknown file extension: %q", extension)
	}

	return err
}
