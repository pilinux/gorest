// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

package lib

import (
	"errors"
	"io/fs"
	"os"
)

// FileExist returns true if the file exists,
// otherwise returns false
func FileExist(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return false
	}

	return true
}
