// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 - 2026 pilinux

package lib

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileExist returns true if the file exists,
// otherwise returns false.
func FileExist(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return false
	}

	return true
}

// ValidatePath validates the given path to prevent directory traversal attacks.
func ValidatePath(fullPath, allowedDir string) (string, error) {
	fullPath = strings.TrimSpace(fullPath)
	allowedDir = strings.TrimSpace(allowedDir)
	if fullPath == "" || allowedDir == "" {
		return "", os.ErrInvalid
	}

	// get absolute path for allowedDir
	absPathAllowedDir, err := filepath.Abs(filepath.Clean(allowedDir))
	if err != nil {
		return "", err
	}

	// get absolute path for fullPath
	absPath, err := filepath.Abs(filepath.Clean(fullPath))
	if err != nil {
		return "", err
	}

	// check if absPath is within absPathAllowedDir
	relPath, err := filepath.Rel(absPathAllowedDir, absPath)
	if err != nil {
		return "", err
	}

	// check directory traversal by looking for ".." in the relative path
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return "", os.ErrInvalid
	}

	return absPath, nil
}
