package lib

import (
	"bytes"
	"image"
	"image/png"
	"os"

	"github.com/google/uuid"
)

// ByteToPNG - generate PNG from bytes and save on the disk
func ByteToPNG(imgByte []byte, path string) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(imgByte))
	if err != nil {
		return "", err
	}

	newImg := "2fa-" + uuid.NewString() + ".png"

	out, _ := os.Create(path + newImg)
	defer out.Close()

	err = png.Encode(out, img)
	if err != nil {
		return "", err
	}

	return newImg, nil
}
