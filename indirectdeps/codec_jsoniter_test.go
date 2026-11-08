//go:build jsoniter

package indirectdeps_test

// Each codec_*_test.go defines wantGinCodec under the same build constraint as
// the matching gin codec/json file; setting two codec tags redeclares it and
// fails to compile.
//
// Indirect modules:
//   - github.com/json-iterator/go
//   - github.com/modern-go/concurrent
//   - github.com/modern-go/reflect2
//
//	Identifier    Purpose
//	------------  -------
//	wantGinCodec  "json.jsoniterApi": TestGinJSONCodec requires %T of ginjson.API to equal it,
//	              so the gin tests ran on json-iterator/go (which uses
//	              modern-go/concurrent and modern-go/reflect2)

const wantGinCodec = "json.jsoniterApi"
