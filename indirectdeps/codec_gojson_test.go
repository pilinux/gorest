//go:build go_json

package indirectdeps_test

// Each codec_*_test.go defines wantGinCodec under the same build constraint as
// the matching gin codec/json file; setting two codec tags redeclares it and
// fails to compile.
//
// Indirect modules:
//   - github.com/goccy/go-json
//
//	Identifier    Purpose
//	------------  -------
//	wantGinCodec  "json.gojsonApi": TestGinJSONCodec requires %T of ginjson.API to equal it,
//	              so the gin tests ran on goccy/go-json

const wantGinCodec = "json.gojsonApi"
