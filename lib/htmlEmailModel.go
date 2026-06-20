package lib

import (
	"strings"
)

// StrArrHTMLModel first slices a string into all substrings separated by `;`
// and then slices each substring separated by `:` for HTMLModel function.
//
// Example: StrArrHTMLModel("key1:value1;key2:value2;key3:value3")
// will return [key1 value1 key2 value2 key3 value3]
func StrArrHTMLModel(s string) []string {
	var out []string

	for seg := range strings.SplitSeq(s, ";") {
		seg = strings.TrimSpace(seg)
		if k, v, ok := strings.Cut(seg, ":"); ok && !strings.Contains(v, ":") {
			out = append(out, strings.TrimSpace(k), strings.TrimSpace(v))
		}
	}

	return out
}

// HTMLModel takes a slice and builds a model for populating
// a templated email.
func HTMLModel(in []string) map[string]any {
	length := len(in)

	model := make(map[string]any)

	for i := 0; i+1 < length; i += 2 {
		model[in[i]] = in[i+1]
	}

	return model
}
