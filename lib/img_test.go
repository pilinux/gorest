package lib_test

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/pilinux/gorest/lib"
)

func TestByteToPNG(t *testing.T) {
	// create a temporary directory to save test images
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// create a test image
	testImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	err = png.Encode(&buf, testImg)
	if err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}

	// generate PNG from bytes and save to disk
	filename, err := lib.ByteToPNG(buf.Bytes(), tempDir+"/")
	if err != nil {
		t.Fatalf("ByteToPNG failed: %v", err)
	}

	// check if the generated file exists
	ok := lib.FileExist(tempDir + "/" + filename)
	if !ok {
		t.Fatalf("failed to find generated file: %v", err)
	}

	// check that the generated file can be decoded as an image
	f, err := os.Open(filepath.Join(tempDir, filename))
	if err != nil {
		t.Fatalf("failed to open generated file: %v", err)
	}
	defer f.Close()

	_, err = png.Decode(f)
	if err != nil {
		t.Fatalf("generated file is not a valid PNG image: %v", err)
	}
}
