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
	defer func() {
		if e := os.RemoveAll(tempDir); e != nil {
			t.Errorf("failed to remove temp dir: %v", e)
		}
	}()

	// create a test image
	testImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	err = png.Encode(&buf, testImg)
	if err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}

	// generate PNG from bytes and save to disk
	filename, err := lib.ByteToPNG(buf.Bytes(), tempDir)
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
	defer func() {
		if e := f.Close(); e != nil {
			t.Errorf("failed to close generated file: %v", e)
		}
	}()

	_, err = png.Decode(f)
	if err != nil {
		t.Fatalf("generated file is not a valid PNG image: %v", err)
	}
}

// TestByteToPNG_InvalidImage tests that ByteToPNG returns an error
// when the input bytes are not a valid image.
func TestByteToPNG_InvalidImage(t *testing.T) {
	_, err := lib.ByteToPNG([]byte("not an image"), t.TempDir())
	if err == nil {
		t.Error("expected error for invalid image bytes, got nil")
	}
}

// TestByteToPNG_EmptyDir tests that ByteToPNG returns an error
// when the directory argument is empty (ValidatePath rejects it).
func TestByteToPNG_EmptyDir(t *testing.T) {
	// create valid PNG bytes
	testImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
	var buf bytes.Buffer
	if err := png.Encode(&buf, testImg); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}

	_, err := lib.ByteToPNG(buf.Bytes(), "")
	if err == nil {
		t.Error("expected error for empty dir, got nil")
	}
}

// TestByteToPNG_CreateFails tests that ByteToPNG returns an error
// when the output file cannot be created (non-existent directory).
func TestByteToPNG_CreateFails(t *testing.T) {
	// create valid PNG bytes
	testImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
	var buf bytes.Buffer
	if err := png.Encode(&buf, testImg); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}

	// use a path that exists for ValidatePath but has a non-existent
	// subdirectory so os.Create fails
	baseDir := t.TempDir()
	noSuchDir := filepath.Join(baseDir, "nonexistent", "subdir")

	_, err := lib.ByteToPNG(buf.Bytes(), noSuchDir)
	if err == nil {
		t.Error("expected error when directory does not exist, got nil")
	}
}
