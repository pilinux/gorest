//go:build !windows

package lib_test

import (
	"bytes"
	"image"
	"image/png"
	"os/signal"
	"syscall"
	"testing"

	"github.com/pilinux/gorest/lib"
)

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
