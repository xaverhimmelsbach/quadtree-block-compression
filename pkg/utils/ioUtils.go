package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
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
	return ReadImageFromReader(file)
}

// ReadImageFromBytes that a byte slice and attempts to decode it into an image
func ReadImageFromBytes(imageBytes []byte) (img image.Image, err error) {
	buffer := bytes.NewBuffer(imageBytes)
	return ReadImageFromReader(buffer)
}

// ReadImageFromReader takes an io.Reader and attempts to decode it into an image
func ReadImageFromReader(reader io.Reader) (img image.Image, err error) {
	img, _, err = image.Decode(reader)
	return img, err
}

// WriteImage encodes an image as either PNG or JPEG according to extension and writes it to writer.
// Note: extension needs to be prepended with a dot.
func WriteImage(img image.Image, writer io.Writer, extension string) (err error) {
	switch extension {
	case ".jpg":
		fallthrough
	case ".jpeg":
		err = jpeg.Encode(writer, img, nil)
	case ".png":
		err = png.Encode(writer, img)
	default:
		err = fmt.Errorf("unknown file extension: %q", extension)
	}

	return err
}

// WriteImageToFile encodes an image as either PNG or JPEG according to the extension of filePath and writes it to filePath.
func WriteImageToFile(img image.Image, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	extension := path.Ext(filePath)

	return WriteImage(img, file, extension)
}

// WriteFile writes data from an io.Reader to filePath
func WriteFile(filePath string, reader io.Reader) error {
	// Create file
	targetFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	// Read buffer
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	// Write bytes
	_, err = targetFile.Write(bytes)
	return err
}
