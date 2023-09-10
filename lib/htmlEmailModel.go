package lib

import (
	"strings"
)

// StrArrHTMLModel first slices a string into all substrings separated by `;`
// and then slices each substring separated by `:` for `HTMLModel` function
//
// Example: StrArrHTMLModel(key1:value1,key2:value2,key3:value3)
// will return [key1 value1 key2 value2 key3 value3]
func StrArrHTMLModel(s string) []string {
	var out []string
	in := strings.Split(s, ";")

	length := len(in)
	for i := 0; i < length; i++ {
		in[i] = strings.TrimSpace(in[i])

		tmp := strings.Split(in[i], ":")
		if len(tmp) == 2 {
			tmp[0] = strings.TrimSpace(tmp[0])
			tmp[1] = strings.TrimSpace(tmp[1])
			out = append(out, tmp[0])
			out = append(out, tmp[1])
		}
	}

	return out
}

// HTMLModel takes a slice and builds a model for populating
// a templated email
func HTMLModel(in []string) map[string]interface{} {
	length := len(in)

	model := make(map[string]interface{})

	for i := 0; i < length; i += 2 {
		model[in[i]] = in[i+1]
	}

	return model
}
