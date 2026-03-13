package lib_test

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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

// TestByteToPNG_EncodeFails tests that ByteToPNG returns an error when
// png.Encode fails due to a write error on the output file.
// It uses RLIMIT_FSIZE to prevent any file write, causing png.Encode
// to fail while os.Create (which produces a zero-byte file) succeeds.
func TestByteToPNG_EncodeFails(t *testing.T) {
	// create valid PNG bytes
	testImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
	var buf bytes.Buffer
	if err := png.Encode(&buf, testImg); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}

	dir := t.TempDir()

	// ignore SIGXFSZ so the process is not killed when the limit is hit
	signal.Ignore(syscall.SIGXFSZ)
	defer signal.Reset(syscall.SIGXFSZ)

	// save and set RLIMIT_FSIZE to 0 to block all file writes
	var orig syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_FSIZE, &orig); err != nil {
		t.Fatalf("failed to get RLIMIT_FSIZE: %v", err)
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 0, Max: orig.Max}); err != nil {
		t.Fatalf("failed to set RLIMIT_FSIZE: %v", err)
	}
	defer func() {
		if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &orig); err != nil {
			t.Errorf("failed to restore RLIMIT_FSIZE: %v", err)
		}
	}()

	_, err := lib.ByteToPNG(buf.Bytes(), dir)
	if err == nil {
		t.Error("expected error when png.Encode cannot write, got nil")
	}
}
