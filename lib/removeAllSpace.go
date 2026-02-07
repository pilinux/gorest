package lib

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import "strings"

// RemoveAllSpace removes all spaces and returns
// the result as a string.
func RemoveAllSpace(s string) string {
	return strings.ReplaceAll(s, " ", "")
}
