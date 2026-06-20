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

// package-level indirections so tests can inject failures
var (
	filepathAbs          = filepath.Abs
	filepathRel          = filepath.Rel
	filepathEvalSymlinks = filepath.EvalSymlinks
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

	// resolve symbolic links before the containment check so that a link
	// inside allowedDir (or its tree) cannot point the path outside the
	// allowed directory while still looking lexically "inside" it
	realAllowedDir, err := resolveSymlinks(absPathAllowedDir)
	if err != nil {
		return "", err
	}
	realPath, err := resolveSymlinks(absPath)
	if err != nil {
		return "", err
	}

	// check if the resolved path is within the resolved allowed directory
	relPath, err := filepathRel(realAllowedDir, realPath)
	if err != nil {
		return "", err
	}

	// check directory traversal by looking for ".." in the relative path
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return "", os.ErrInvalid
	}

	// NOTE: the returned path is validated but opened separately by the
	// caller, leaving a residual TOCTOU window if a path component is
	// replaced with a symlink between validation and use.
	return absPath, nil
}

// resolveSymlinks returns path with the symbolic links in its existing
// components resolved. The final target may not exist yet (e.g. a file about
// to be created), so it resolves the deepest existing ancestor and re-attaches
// the remaining, not-yet-existing components.
func resolveSymlinks(path string) (string, error) {
	resolved, err := filepathEvalSymlinks(path)
	if err == nil {
		return resolved, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}

	parent := filepath.Dir(path)
	if parent == path {
		// reached the filesystem root; nothing left to resolve
		return path, nil
	}

	resolvedParent, err := resolveSymlinks(parent)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedParent, filepath.Base(path)), nil
}
