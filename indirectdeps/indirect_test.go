// Package indirectdeps_test pins how gorest uses each indirect dependency,
// so bumping one for a security fix shows quickly whether our code paths still work.
package indirectdeps_test

// Coverage map: ties every indirect requirement in go.mod to the gorest path
// that reaches it and to the tests that exercise that path.
//
// Indirect modules:
//   - all indirect requirements in go.mod, one indirectUses entry each
//
//	Identifier                      Purpose
//	------------------------------  -------
//	indirectUse                     via: direct dependency that pulls the module in; usage: how gorest
//	                                reaches it, or "not reached"/"not linked" plus why; tests
//	ginJSONTests, sonicUsage        shared values for the codec-tag entries
//	indirectUses                    one entry per indirect module in go.mod
//	indirectRequirements            parses go.mod, returns the modules marked indirect
//	declaredTests                   parses indirectdeps/*_test.go (every build tag), returns Test* names
//	TestIndirectDependenciesMapped  fails when go.mod has an unmapped indirect module, a mapped
//	                                module is no longer indirect, tests are listed for an
//	                                unreached module or missing for a reached one, or a listed
//	                                test does not exist

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type indirectUse struct {
	via   string   // direct dependency that pulls it in
	usage string   // how gorest reaches it; "not reached"/"not linked" if never
	tests []string // tests that exercise that usage
}

var (
	ginJSONTests = []string{"TestGinJSONCodec", "TestGinJSONBindingAndRender"}
	sonicUsage   = "gin JSON codec, only with -tags=sonic"
)

// indirectUses must list every "// indirect" requirement in go.mod.
var indirectUses = map[string]indirectUse{
	"filippo.io/edwards25519": {"go-sql-driver/mysql", "MariaDB client_ed25519 auth in database.InitDB", []string{"TestMySQLEd25519Auth", "TestMySQLEd25519AuthWrongPassword"}},

	"github.com/bytedance/gopkg":            {"gin-gonic/gin", sonicUsage, ginJSONTests},
	"github.com/bytedance/sonic":            {"gin-gonic/gin", sonicUsage, ginJSONTests},
	"github.com/bytedance/sonic/loader":     {"gin-gonic/gin", sonicUsage, ginJSONTests},
	"github.com/cloudwego/base64x":          {"gin-gonic/gin", sonicUsage, ginJSONTests},
	"github.com/klauspost/cpuid/v2":         {"gin-gonic/gin", sonicUsage, ginJSONTests},
	"github.com/twitchyliquid64/golang-asm": {"gin-gonic/gin", sonicUsage + " (amd64 JIT)", ginJSONTests},
	"golang.org/x/arch":                     {"gin-gonic/gin", sonicUsage + " (amd64 JIT)", ginJSONTests},
	"github.com/goccy/go-json":              {"gin-gonic/gin", "gin JSON codec, only with -tags=go_json", ginJSONTests},
	"github.com/json-iterator/go":           {"gin-gonic/gin", "gin JSON codec, only with -tags=jsoniter", ginJSONTests},
	"github.com/modern-go/concurrent":       {"json-iterator/go", "gin JSON codec, only with -tags=jsoniter", ginJSONTests},
	"github.com/modern-go/reflect2":         {"json-iterator/go", "gin JSON codec, only with -tags=jsoniter", ginJSONTests},

	"github.com/go-playground/validator/v10":        {"gin-gonic/gin", "runs on every ShouldBindJSON", []string{"TestGinValidator", "TestGinJSONBindingAndRender"}},
	"github.com/go-playground/locales":              {"validator", "not reached: validator translations are never registered", nil},
	"github.com/go-playground/universal-translator": {"validator", "not reached: validator translations are never registered", nil},
	"github.com/gabriel-vasile/mimetype":            {"validator", "not reached: no `mimetype` validation tags", nil},
	"github.com/leodido/go-urn":                     {"validator", "not reached: no `urn_rfc2141` validation tags", nil},
	"github.com/gin-contrib/sse":                    {"gin-gonic/gin", "not reached: no c.SSEvent / c.Stream", nil},
	"github.com/goccy/go-yaml":                      {"gin-gonic/gin", "not reached: only JSON binding/rendering is used", nil},
	"github.com/pelletier/go-toml/v2":               {"gin-gonic/gin", "not reached: only JSON binding/rendering is used", nil},
	"github.com/ugorji/go/codec":                    {"gin-gonic/gin", "not reached: only JSON binding/rendering is used", nil},
	"google.golang.org/protobuf":                    {"gin-gonic/gin", "not reached: only JSON binding/rendering is used", nil},
	"github.com/quic-go/quic-go":                    {"gin-gonic/gin", "not reached: no engine.RunQUIC", nil},
	"github.com/quic-go/qpack":                      {"quic-go", "not reached: no engine.RunQUIC", nil},
	"github.com/mattn/go-isatty":                    {"gin-gonic/gin", "gin logger color detection", []string{"TestGinLoggerNonTTY"}},

	"github.com/jackc/pgx/v5":        {"gorm.io/driver/postgres", "pgx stdlib driver, DSN parsing, SCRAM in database.InitDB", []string{"TestPostgresSCRAMPrecis", "TestPostgresSCRAMWrongPassword", "TestPostgresPassfile", "TestPostgresServiceFile"}},
	"github.com/jackc/pgpassfile":    {"pgx", "password lookup when DBPASS is empty", []string{"TestPostgresPassfile"}},
	"github.com/jackc/pgservicefile": {"pgx", "PGSERVICE / PGSERVICEFILE settings", []string{"TestPostgresServiceFile"}},
	"github.com/jackc/puddle/v2":     {"pgx", "not reached: pgxpool is unused, database/sql pools instead", nil},

	"github.com/jinzhu/inflection": {"gorm.io/gorm", "pluralized table names in migrate.StartMigration", []string{"TestSQLiteMigrateAndCRUD"}},
	"github.com/jinzhu/now":        {"gorm.io/gorm", "not reached: no time `default`/`zerovalue` tags or string time assignment", nil},
	"github.com/mattn/go-sqlite3":  {"gorm.io/driver/sqlite", "sqlite3 driver in database.InitDB", []string{"TestSQLiteMigrateAndCRUD", "TestGormJoinTableNaming"}},

	"github.com/tilinna/clock": {"mediocregopher/radix/v4", "pool timing and periodic PING in database.InitRedis", []string{"TestRedisPoolAndCommands"}},

	"github.com/klauspost/compress": {"mongo-driver/v2", "snappy/zstd wire compression via MONGO_URI", []string{"TestMongoWireCompression"}},
	"github.com/xdg-go/scram":       {"mongo-driver/v2", "SCRAM-SHA-1/256 auth via MONGO_URI", []string{"TestMongoSCRAM"}},
	"github.com/xdg-go/stringprep":  {"mongo-driver/v2", "SASLprep of SCRAM-SHA-256 passwords", []string{"TestMongoSCRAM"}},
	"github.com/youmark/pkcs8":      {"mongo-driver/v2", "encrypted tlsCertificateKeyFile via MONGO_URI", []string{"TestMongoTLSEncryptedKeyAndOCSP"}},

	"github.com/pilinux/cryptoengine": {"pilinux/twofactor", "sealing TOTP state in lib.NewTOTP / ValidateTOTP", []string{"TestTwoFactorTOTP"}},
	"github.com/sec51/convert":        {"pilinux/twofactor", "TOTP time-step rounding, big-endian state encoding", []string{"TestTwoFactorTOTP"}},
	"rsc.io/qr":                       {"pilinux/twofactor", "enrolment QR in lib.NewQR", []string{"TestTwoFactorTOTP"}},

	"github.com/pkg/errors": {"ulule/limiter/v3", "rate format errors from lib.InitRateLimiter", []string{"TestRateLimiter"}},

	"golang.org/x/net":  {"gin-gonic/gin", "not reached: no engine.UseH2C, no RunQUIC", nil},
	"golang.org/x/sync": {"mongo-driver/v2", "OCSP responder fan-out on TLS connections", []string{"TestMongoTLSEncryptedKeyAndOCSP"}},
	"golang.org/x/sys": {"logrus, gin, argon2, sentry-go", "terminal ioctl, CPU feature detection, execabs", []string{
		"TestGinLoggerNonTTY", "TestLogrusNonTTY", "TestArgon2Hashing", "TestSentryReleaseAndRateLimit"}},
	"golang.org/x/text": {"gorm, pgx, xdg-go/stringprep, sentry-go", "gorm naming cases, pgx precis, SASLprep norm (sentry-go: debug logging only)", []string{
		"TestGormJoinTableNaming", "TestPostgresSCRAMPrecis", "TestMongoSCRAM"}},

	"github.com/go-errors/errors":     {"getsentry/sentry-go", "not linked: sentry-go test dependency", nil},
	"github.com/rogpeppe/go-internal": {"flosch/pongo2", "not linked: pongo2 test dependency", nil},
}

// indirectRequirements returns each go.mod module marked indirect with its version.
func indirectRequirements(t *testing.T) map[string]string {
	t.Helper()
	f, err := os.Open(filepath.Join(repoRoot, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	mods := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line, _ := strings.CutPrefix(strings.TrimSpace(sc.Text()), "require ")
		if fields := strings.Fields(line); len(fields) >= 2 && strings.HasSuffix(line, "// indirect") {
			mods[fields[0]] = fields[1]
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return mods
}

func declaredTests(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(repoRoot, "indirectdeps", "*_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	fset := token.NewFileSet()
	for _, name := range files {
		f, err := parser.ParseFile(fset, name, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv == nil && strings.HasPrefix(fn.Name.Name, "Test") {
				names = append(names, fn.Name.Name)
			}
		}
	}
	return names
}

// TestIndirectDependenciesMapped fails when go.mod gains an indirect
// dependency nobody has traced yet, or when a mapping goes stale.
func TestIndirectDependenciesMapped(t *testing.T) {
	mods := indirectRequirements(t)
	if len(mods) == 0 {
		t.Fatal("no indirect requirements parsed from go.mod")
	}
	for m := range mods {
		if _, ok := indirectUses[m]; !ok {
			t.Errorf("%s is a new indirect dependency: trace its use (go mod why -m %s) and add it to indirectUses", m, m)
		}
	}

	tests := declaredTests(t)
	for m, use := range indirectUses {
		if _, ok := mods[m]; !ok {
			t.Errorf("%s is no longer an indirect requirement in go.mod; drop it from indirectUses", m)
		}
		unreached := strings.HasPrefix(use.usage, "not reached") || strings.HasPrefix(use.usage, "not linked")
		if unreached != (len(use.tests) == 0) {
			t.Errorf("%s: tests must be listed exactly when the dependency is reached", m)
		}
		for _, name := range use.tests {
			if !slices.Contains(tests, name) {
				t.Errorf("%s: test %s does not exist", m, name)
			}
		}
	}
}
