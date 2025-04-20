// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

package lib

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileExist returns true if the file exists,
// otherwise returns false
func FileExist(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return false
	}

	return true
}

// ValidatePath validates the given path to prevent directory traversal attacks
func ValidatePath(fullPath, allowedDir string) (string, error) {
	// clean and get absolute path of the given fullPath
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}

	// check for directory traversal patterns
	if strings.Contains(absPath, "../") || strings.Contains(absPath, "..\\") {
		return "", os.ErrInvalid
	}

	// clean and get absolute path of the allowedDir
	// this is the directory where the file should be saved
	absPathAllowedDir, err := filepath.Abs(allowedDir)
	if err != nil {
		return "", err
	}

	// ensure the absPath is within the allowed directory
	if !strings.HasPrefix(absPath, filepath.Clean(absPathAllowedDir)) {
		return "", os.ErrInvalid
	}

	return absPath, nil
}
