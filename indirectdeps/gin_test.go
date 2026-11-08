package indirectdeps_test

// gin path: controllers bind with c.ShouldBindJSON and reply through
// renderer.Render (c.SecureJSON); both go through gin's codec/json and
// validator.
//
// Indirect modules:
//   - github.com/go-playground/validator/v10
//   - github.com/mattn/go-isatty
//   - golang.org/x/sys
//   - codec_*_test.go modules, only under a codec tag
//
//	Function                     Purpose
//	---------------------------  -------
//	init                         gin.SetMode(gin.TestMode) for the package
//	doRequest                    serves one JSON request through a gin engine, returns the recorder
//	TestGinJSONCodec             the compiled codec equals wantGinCodec, and model.AuthPayload with
//	                             \u00fc, a quote and <>& round-trips through ginjson.API.
//	                             Covers the sonic, go-json and jsoniter modules under their tags
//	TestGinJSONBindingAndRender  POST /auth binds model.Auth (UnmarshalJSON hashes the password
//	                             with argon2id): the hash verifies and the response has no
//	                             password. POST /payload renders a slice: SecureJSON prefixes
//	                             while(1);. Malformed JSON, an array, a wrong field type and a
//	                             short password give 400. Covers the codec modules and
//	                             go-playground/validator/v10
//	TestGinValidator             binding.Validator has an engine (validator/v10) and ValidateStruct
//	                             accepts Auth, AuthPayload, TempEmail, TwoFA and []AuthPayload
//	TestGinLoggerNonTTY          gin.LoggerWithConfig to a file with TERM=xterm-256color: the line
//	                             contains /ping and no ANSI escape. Covers mattn/go-isatty and
//	                             golang.org/x/sys (isatty's ioctl)
//	TestLogrusNonTTY             logrus TextFormatter to a file: level=error msg=boom, no ANSI
//	                             escape. Covers golang.org/x/sys/unix (logrus' terminal check)

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	ginjson "github.com/gin-gonic/gin/codec/json"
	"github.com/pilinux/argon2"
	log "github.com/sirupsen/logrus"

	gmodel "github.com/pilinux/gorest/database/model"
	glib "github.com/pilinux/gorest/lib"
	grenderer "github.com/pilinux/gorest/lib/renderer"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func doRequest(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// TestGinJSONCodec pins which codec gin compiled in; run with -tags=sonic,
// -tags=go_json or -tags=jsoniter to cover those indirect dependencies.
func TestGinJSONCodec(t *testing.T) {
	logIndirect(t)
	if got := fmt.Sprintf("%T", ginjson.API); got != wantGinCodec {
		t.Fatalf("gin JSON codec = %s, want %s", got, wantGinCodec)
	}

	in := gmodel.AuthPayload{Email: "a@b.c", OTP: "123456", PassNew: "\u00fc\"<>&"} // ggignore
	b, err := ginjson.API.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out gmodel.AuthPayload
	if err := ginjson.API.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Errorf("round trip = %+v, want %+v", out, in)
	}
}

// TestGinJSONBindingAndRender replays the controller pattern:
// ShouldBindJSON into our models, then renderer.Render (SecureJSON / abort).
func TestGinJSONBindingAndRender(t *testing.T) {
	logIndirect(t)
	// model.Auth hashes the password with argon2id while unmarshalling
	cfg := setupConfig(t)
	cfg.Security.UserPassMinLength = 6
	cfg.Security.HashPass = glib.HashPassConfig{Memory: 64, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}
	cfg.Security.HashSec = "pepper" // ggignore

	var bound gmodel.Auth
	r := gin.New()
	r.POST("/auth", func(c *gin.Context) {
		auth := gmodel.Auth{}
		if err := c.ShouldBindJSON(&auth); err != nil {
			grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
			return
		}
		bound = auth
		grenderer.Render(c, auth, http.StatusCreated)
	})
	r.POST("/payload", func(c *gin.Context) {
		payload := gmodel.AuthPayload{}
		if err := c.ShouldBindJSON(&payload); err != nil {
			grenderer.Render(c, gin.H{"message": "bad"}, http.StatusBadRequest)
			return
		}
		grenderer.Render(c, []gmodel.AuthPayload{payload}, http.StatusOK)
	})

	w := doRequest(r, http.MethodPost, "/auth", `{"authID":7,"email":"a@b.c","password":"secret"}`) // ggignore
	if w.Code != http.StatusCreated {
		t.Fatalf("POST /auth = %d: %s", w.Code, w.Body)
	}
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("response is not JSON: %v: %s", err, w.Body)
	}
	if got["email"] != "a@b.c" || got["authID"] != float64(7) {
		t.Errorf("auth response = %v", got)
	}
	if _, leaked := got["password"]; leaked {
		t.Error("password leaked in response")
	}
	if ok, err := argon2.ComparePasswordAndHash("secret", "pepper", bound.Password); !ok || err != nil { // ggignore
		t.Errorf("bound password is not an argon2id hash of the input: %v", err)
	}

	// SecureJSON prefixes top-level arrays to defeat JSON hijacking
	w = doRequest(r, http.MethodPost, "/payload", `{"email":"x@y.z","otp":"654321"}`) // ggignore
	body, ok := strings.CutPrefix(w.Body.String(), "while(1);")
	if w.Code != http.StatusOK || !ok {
		t.Fatalf("POST /payload = %d %q", w.Code, w.Body)
	}
	var list []gmodel.AuthPayload
	if err := json.Unmarshal([]byte(body), &list); err != nil || len(list) != 1 || list[0].OTP != "654321" { // ggignore
		t.Errorf("payload list = %+v, %v", list, err)
	}

	for _, bad := range []string{`{"email":`, `[1,2]`, `{"authID":"seven"}`, `{"email":"a@b.c","password":"short"}`} { // ggignore
		if w := doRequest(r, http.MethodPost, "/auth", bad); w.Code != http.StatusBadRequest {
			t.Errorf("POST /auth %q = %d, want 400", bad, w.Code)
		}
	}
}

// TestGinValidator covers go-playground/validator/v10: gin runs it on every
// ShouldBindJSON, so it must accept our tag-less request models.
func TestGinValidator(t *testing.T) {
	logIndirect(t)
	if binding.Validator.Engine() == nil {
		t.Fatal("gin has no validator engine")
	}
	for _, obj := range []any{
		&gmodel.Auth{Email: "a@b.c"},
		&gmodel.AuthPayload{OTP: "1"}, // ggignore
		&gmodel.TempEmail{Email: "a@b.c"},
		&gmodel.TwoFA{},
		[]gmodel.AuthPayload{{}, {}},
	} {
		if err := binding.Validator.ValidateStruct(obj); err != nil {
			t.Errorf("ValidateStruct(%T) = %v", obj, err)
		}
	}
}

// TestGinLoggerNonTTY covers mattn/go-isatty (and golang.org/x/sys): gin's
// logger checks the output fd and must not emit colors to a plain file.
func TestGinLoggerNonTTY(t *testing.T) {
	logIndirect(t)
	t.Setenv("TERM", "xterm-256color")
	f, err := os.CreateTemp(t.TempDir(), "gin-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: f}))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	doRequest(r, http.MethodGet, "/ping", "")

	out, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "/ping") {
		t.Fatalf("access log missing request: %q", out)
	}
	if strings.Contains(string(out), "\x1b[") {
		t.Errorf("colored output written to a non-terminal: %q", out)
	}
}

// TestLogrusNonTTY covers golang.org/x/sys/unix: logrus' TextFormatter (our
// default logger) probes the fd with an ioctl to decide on colors.
func TestLogrusNonTTY(t *testing.T) {
	logIndirect(t)
	f, err := os.CreateTemp(t.TempDir(), "logrus-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	logger := log.New()
	logger.Out = f
	logger.WithField("code", "1001.1").Error("boom")

	out, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if s := string(out); !strings.Contains(s, "level=error") || !strings.Contains(s, `msg=boom`) || strings.Contains(s, "\x1b[") {
		t.Errorf("unexpected non-TTY log line: %q", s)
	}
}
