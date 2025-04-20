package lib

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// ByteToPNG - generate PNG from bytes and save on the disk
func ByteToPNG(imgByte []byte, dir string) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(imgByte))
	if err != nil {
		return "", err
	}

	newImg := "2fa-" + uuid.NewString() + ".png"
	fullPath := filepath.Join(dir, newImg)

	// prevent directory traversal attacks by validating the path
	fullPath, err = ValidatePath(fullPath, dir)
	if err != nil {
		return "", err
	}

	out, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer func() {
		if e := out.Close(); e != nil && err == nil {
			err = e
		}
	}()

	err = png.Encode(out, img)
	if err != nil {
		return "", err
	}

	return newImg, nil
}
