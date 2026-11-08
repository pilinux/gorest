//go:build sonic && (linux || windows || darwin)

package indirectdeps_test

// Each codec_*_test.go defines wantGinCodec under the same build constraint as
// the matching gin codec/json file; setting two codec tags redeclares it and
// fails to compile.
//
// Indirect modules:
//   - github.com/bytedance/sonic
//   - github.com/bytedance/sonic/loader
//   - github.com/bytedance/gopkg
//   - github.com/cloudwego/base64x
//   - github.com/klauspost/cpuid/v2
//   - github.com/twitchyliquid64/golang-asm
//   - golang.org/x/arch
//
//	Identifier    Purpose
//	------------  -------
//	wantGinCodec  "json.sonicApi": TestGinJSONCodec requires %T of ginjson.API to equal it,
//	              so the gin tests ran on bytedance/sonic. amd64 links sonic,
//	              sonic/loader, gopkg, base64x, cpuid, golang-asm and x/arch;
//	              arm64 links sonic, gopkg and x/arch (no JIT)

const wantGinCodec = "json.sonicApi"
