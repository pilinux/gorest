package lib_test

import (
	"os"
	"strings"
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

// Test function for validatePath
func TestValidatePath(t *testing.T) {
	// set up a valid directory for testing
	allowedDir := "/home/user/allowed_dir"

	tests := []struct {
		fullPath       string
		expectedResult string
		expectedErr    error
	}{
		{
			fullPath:       "/home/user/allowed_dir/file.txt", // valid path inside the allowed directory
			expectedResult: "/home/user/allowed_dir/file.txt",
			expectedErr:    nil,
		},
		{
			fullPath:       "/home/user/allowed_dir/../evil.txt", // invalid path with directory traversal
			expectedResult: "",
			expectedErr:    os.ErrInvalid,
		},
		{
			fullPath:       "/home/user/allowed_dir/..\\evil.txt", // Windows-style directory traversal
			expectedResult: "",
			expectedErr:    os.ErrInvalid,
		},
		{
			fullPath:       "/home/user/other_dir/file.txt", // path outside the allowed directory
			expectedResult: "",
			expectedErr:    os.ErrInvalid,
		},
		{
			fullPath:       "/home/user/allowed_dir/./file.txt", // valid path with redundant './'
			expectedResult: "/home/user/allowed_dir/file.txt",
			expectedErr:    nil,
		},
		{
			fullPath:       "/home/user/allowed_dir/subdir/../file.txt", // valid path with ../ inside allowed dir
			expectedResult: "/home/user/allowed_dir/file.txt",
			expectedErr:    nil,
		},
		{
			fullPath:       "/home/user/allowed_dir/../../other_dir/file.txt", // invalid path outside allowed dir
			expectedResult: "",
			expectedErr:    os.ErrInvalid,
		},
		{
			fullPath:       "", // empty path
			expectedResult: "",
			expectedErr:    os.ErrInvalid,
		},
	}

	// loop through all test cases
	for _, tt := range tests {
		t.Run(tt.fullPath, func(t *testing.T) {
			// run validatePath function
			result, err := lib.ValidatePath(tt.fullPath, allowedDir)

			// compare results
			if result != tt.expectedResult {
				t.Errorf("expected result '%s', got '%s'", tt.expectedResult, result)
			}

			// check for error message comparison
			if err != nil && tt.expectedErr != nil {
				if !strings.Contains(err.Error(), tt.expectedErr.Error()) {
					t.Errorf("expected error '%v', got '%v'", tt.expectedErr, err)
				}
			} else if err != nil && tt.expectedErr == nil {
				t.Errorf("expected no error, got '%v'", err)
			} else if err == nil && tt.expectedErr != nil {
				t.Errorf("expected error '%v', got no error", tt.expectedErr)
			}
		})
	}
}
