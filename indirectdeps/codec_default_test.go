//go:build !jsoniter && !go_json && !(sonic && (linux || windows || darwin))

package indirectdeps_test

// Each codec_*_test.go defines wantGinCodec under the same build constraint as
// the matching gin codec/json file; setting two codec tags redeclares it and
// fails to compile.
//
// Indirect modules:
//   - none (encoding/json)
//
//	Identifier    Purpose
//	------------  -------
//	wantGinCodec  "json.jsonApi": TestGinJSONCodec requires %T of ginjson.API to equal it,
//	              so the gin tests ran on encoding/json (no indirect module)

const wantGinCodec = "json.jsonApi"
