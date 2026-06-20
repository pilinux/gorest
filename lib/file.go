// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 - 2026 pilinux

package lib

import (
	"os"
	"path/filepath"
	"strings"
)

// package-level indirections so tests can inject failures
var (
	filepathAbs = filepath.Abs
	filepathRel = filepath.Rel
)

// FileExist returns true only if the file exists and is stat-able.
// Any error (including ambiguous ones such as permission or I/O errors)
// yields false, since existence cannot be confirmed.
func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ValidatePath validates the given path to prevent directory traversal attacks.
func ValidatePath(fullPath, allowedDir string) (string, error) {
	fullPath = strings.TrimSpace(fullPath)
	allowedDir = strings.TrimSpace(allowedDir)
	if fullPath == "" || allowedDir == "" {
		return "", os.ErrInvalid
	}

	// get absolute path for allowedDir
	absPathAllowedDir, err := filepathAbs(filepath.Clean(allowedDir))
	if err != nil {
		return "", err
	}

	// get absolute path for fullPath
	absPath, err := filepathAbs(filepath.Clean(fullPath))
	if err != nil {
		return "", err
	}

	// check if absPath is within absPathAllowedDir
	relPath, err := filepathRel(absPathAllowedDir, absPath)
	if err != nil {
		return "", err
	}

	// check directory traversal by looking for ".." in the relative path
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return "", os.ErrInvalid
	}

	return absPath, nil
}
