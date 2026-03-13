package lib

import "path/filepath"

// SetFilepathAbs replaces the internal filepathAbs function and
// returns a restore function. Intended for use in tests only.
func SetFilepathAbs(fn func(string) (string, error)) func() {
	orig := filepathAbs
	filepathAbs = fn
	return func() { filepathAbs = orig }
}

// SetFilepathRel replaces the internal filepathRel function and
// returns a restore function. Intended for use in tests only.
func SetFilepathRel(fn func(string, string) (string, error)) func() {
	orig := filepathRel
	filepathRel = fn
	return func() { filepathRel = orig }
}

// ResetFilepathFuncs restores all filepath indirections to their defaults.
func ResetFilepathFuncs() {
	filepathAbs = filepath.Abs
	filepathRel = filepath.Rel
}
