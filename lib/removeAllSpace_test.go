package lib_test

import (
	"testing"

	"github.com/pilinux/gorest/lib"
)

func TestRemoveAllSpace(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{"Hello, World!", "Hello,World!"},
		{"    This   string   has   lots   of   spaces   ", "Thisstringhaslotsofspaces"},
		{"No spaces here", "Nospaceshere"},
		{"     ", ""},
		{"", ""},
	}

	for _, tc := range testCases {
		got := lib.RemoveAllSpace(tc.input)
		if got != tc.want {
			t.Errorf("lib.RemoveAllSpace(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
