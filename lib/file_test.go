package lib_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pilinux/gorest/lib"
)

func TestFileExist(t *testing.T) {
	// create temporary file
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		if e := os.Remove(tempFile.Name()); e != nil {
			t.Errorf("failed to remove temp file: %v", e)
		}
	}()

	// test with existing file
	if !lib.FileExist(tempFile.Name()) {
		t.Error("expected existing file to return true, but got false")
	}

	// test with non-existing file
	if lib.FileExist("non-existing-file") {
		t.Error("expected non-existing file to return false, but got true")
	}
}

// TestValidatePath tests the ValidatePath function.
func TestValidatePath(t *testing.T) {
	baseDir := t.TempDir()
	allowedDir := filepath.Join(baseDir, "allowed")
	if err := os.MkdirAll(allowedDir, 0750); err != nil {
		t.Fatalf("failed to create allowed dir: %v", err)
	}

	tests := []struct {
		name           string
		fullPath       string
		allowedDir     string
		expectedResult string
		expectedErr    error
	}{
		{
			name:           "absolute path inside allowed directory",
			fullPath:       filepath.Join(allowedDir, "file.txt"),
			allowedDir:     allowedDir,
			expectedResult: filepath.Join(allowedDir, "file.txt"),
			expectedErr:    nil,
		},
		{
			name:           "escaped path rejected",
			fullPath:       filepath.Join(allowedDir, "..", "evil.txt"),
			allowedDir:     allowedDir,
			expectedResult: "",
			expectedErr:    os.ErrInvalid,
		},
		{
			name:           "shared prefix path rejected",
			fullPath:       filepath.Join(baseDir, "allowed-evil", "file.txt"),
			allowedDir:     allowedDir,
			expectedResult: "",
			expectedErr:    os.ErrInvalid,
		},
		{
			name:           "redundant current dir resolves inside allowed directory",
			fullPath:       filepath.Join(allowedDir, ".", "file.txt"),
			allowedDir:     allowedDir,
			expectedResult: filepath.Join(allowedDir, "file.txt"),
			expectedErr:    nil,
		},
		{
			name:           "parent segment inside allowed directory resolves valid",
			fullPath:       filepath.Join(allowedDir, "subdir", "..", "file.txt"),
			allowedDir:     allowedDir,
			expectedResult: filepath.Join(allowedDir, "file.txt"),
			expectedErr:    nil,
		},
		{
			name:           "empty path rejected",
			fullPath:       "",
			allowedDir:     allowedDir,
			expectedResult: "",
			expectedErr:    os.ErrInvalid,
		},
		{
			name:           "empty allowed dir rejected",
			fullPath:       filepath.Join(allowedDir, "file.txt"),
			allowedDir:     "",
			expectedResult: "",
			expectedErr:    os.ErrInvalid,
		},
	}

	// loop through all test cases
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			// run validatePath function
			result, err := lib.ValidatePath(tt.fullPath, tt.allowedDir)

			// compare results
			if result != tt.expectedResult {
				t.Errorf("expected result '%s', got '%s'", tt.expectedResult, result)
			}

			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error '%v', got '%v'", tt.expectedErr, err)
				}
			} else if err != nil && tt.expectedErr == nil {
				t.Errorf("expected no error, got '%v'", err)
			}
		})
	}
}

// TestValidatePath_AbsAllowedDirError tests that ValidatePath returns an
// error when filepath.Abs fails for the allowedDir argument.
func TestValidatePath_AbsAllowedDirError(t *testing.T) {
	errAbs := fmt.Errorf("abs error")
	restore := lib.SetFilepathAbs(func(_ string) (string, error) {
		return "", errAbs
	})
	defer restore()

	result, err := lib.ValidatePath("/some/path", "/allowed")
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
	if !errors.Is(err, errAbs) {
		t.Errorf("expected abs error, got %v", err)
	}
}

// TestValidatePath_AbsFullPathError tests that ValidatePath returns an
// error when filepath.Abs fails for the fullPath argument (second call).
func TestValidatePath_AbsFullPathError(t *testing.T) {
	errAbs := fmt.Errorf("abs fullPath error")
	callCount := 0
	restore := lib.SetFilepathAbs(func(path string) (string, error) {
		callCount++
		if callCount == 1 {
			// first call (allowedDir) succeeds
			return filepath.Abs(path)
		}
		// second call (fullPath) fails
		return "", errAbs
	})
	defer restore()

	result, err := lib.ValidatePath("/some/path", "/allowed")
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
	if !errors.Is(err, errAbs) {
		t.Errorf("expected abs fullPath error, got %v", err)
	}
}

// TestValidatePath_RelError tests that ValidatePath returns an error
// when filepath.Rel fails.
func TestValidatePath_RelError(t *testing.T) {
	errRel := fmt.Errorf("rel error")
	restore := lib.SetFilepathRel(func(_, _ string) (string, error) {
		return "", errRel
	})
	defer restore()

	result, err := lib.ValidatePath("/allowed/file.txt", "/allowed")
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
	if !errors.Is(err, errRel) {
		t.Errorf("expected rel error, got %v", err)
	}
}
