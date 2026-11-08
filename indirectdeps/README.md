# indirectdeps

Regression tests for the indirect dependencies in `go.mod`. Each test drives
gorest's own entry points (`database.InitDB`, `InitRedis`, `InitMongo`,
`migrate`, `lib`, `middleware`, `renderer`) through the exact code path that
reaches an indirect module. Run them after bumping an indirect dependency
(e.g. for a vulnerability fix) to confirm nothing broke.

No live services or env vars are needed: MySQL, PostgreSQL, Redis, MongoDB,
Sentry and OCSP are in-process fakes. SQLite needs cgo.

```bash
go test ./indirectdeps/...
# gin's alternative JSON codecs pull in their own indirect modules
go test -tags=sonic ./indirectdeps/...
go test -tags=go_json ./indirectdeps/...
go test -tags=jsoniter ./indirectdeps/...
```

With `-v`, each test prints the indirect modules it uses that are linked
into the test binary, with the versions from its build info:

```txt
[indirectdeps] TestMongoSCRAM uses:
  - github.com/xdg-go/scram v1.4.0
  - github.com/xdg-go/stringprep v1.0.4
  - golang.org/x/text v0.43.0
```

## JSON codec tags

Each tag swaps the library behind gin's `ShouldBindJSON` and rendering; the
tests stay the same, only the compiled indirect modules change.

| Tag | gin uses | Indirect modules exercised |
| --- | --- | --- |
| none | `encoding/json` | none (what the examples ship with) |
| `sonic` | `bytedance/sonic` | amd64: `sonic`, `sonic/loader`, `gopkg`, `cloudwego/base64x`, `klauspost/cpuid`, `golang-asm`, `x/arch`; arm64: `sonic`, `gopkg`, `x/arch` (no JIT) |
| `go_json` | `goccy/go-json` | `goccy/go-json` |
| `jsoniter` | `json-iterator/go` | `json-iterator/go`, `modern-go/concurrent`, `modern-go/reflect2` |

- Use one tag per run; `go_json,jsoniter` together does not compile.
- Go ignores unknown tags, so a typo silently tests `encoding/json`.
- `sonic` works only on linux, windows and darwin. Run it on linux/amd64 to
  cover the JIT modules.
- `TestGinJSONCodec` asserts the expected codec is active (`codec_*_test.go`),
  so a tag that did not take effect fails.
- A failure under one tag only matters if you build the app with that tag.

## Coverage map

`indirect_test.go` maps every `// indirect` requirement to the gorest code
path that reaches it and to the tests covering that path. Modules that are
linked but never reached (or only used by upstream tests) are listed with a
reason. `TestIndirectDependenciesMapped` fails when `go.mod` gains an
indirect module that is not mapped yet; trace it with `go mod why -m <module>`.

The tests never import an indirect module directly, so `go mod tidy` keeps
them marked `// indirect`.
