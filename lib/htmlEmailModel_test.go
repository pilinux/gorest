package lib_test

import (
	"reflect"
	"testing"

	"github.com/pilinux/gorest/lib"
)

func TestStrArrHTMLModel(t *testing.T) {
	testCases := []struct {
		input string
		want  []string
	}{
		{
			"title: My Page; body: Welcome to my page!",
			[]string{"title", "My Page", "body", "Welcome to my page!"},
		},
		{
			"title: My Page",
			[]string{"title", "My Page"},
		},
		{
			"title: My Page; body: Welcome to my page!; footer: Copyright 2023",
			[]string{"title", "My Page", "body", "Welcome to my page!", "footer", "Copyright 2023"},
		},
	}

	for _, tc := range testCases {
		got := lib.StrArrHTMLModel(tc.input)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("lib.StrArrHTMLModel(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestHTMLModel(t *testing.T) {
	testCases := []struct {
		input []string
		want  map[string]interface{}
	}{
		{
			[]string{"title", "My Page", "body", "Welcome to my page!"},
			map[string]interface{}{"title": "My Page", "body": "Welcome to my page!"},
		},
		{
			[]string{"title", "My Page"},
			map[string]interface{}{"title": "My Page"},
		},
		{
			[]string{"title", "My Page", "body", "Welcome to my page!", "footer", "Copyright 2023"},
			map[string]interface{}{"title": "My Page", "body": "Welcome to my page!", "footer": "Copyright 2023"},
		},
		{
			[]string{},
			map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		got := lib.HTMLModel(tc.input)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("lib.HTMLModel(%v) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
