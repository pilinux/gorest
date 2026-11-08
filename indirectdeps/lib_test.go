package indirectdeps_test

// lib paths: 2FA (handler/twoFA.go, service/common.go), password hashing and
// the rate limiter.
//
// Indirect modules:
//   - github.com/pilinux/cryptoengine
//   - github.com/sec51/convert
//   - rsc.io/qr
//   - golang.org/x/sys
//   - github.com/pkg/errors
//
//	Function              Purpose
//	--------------------  -------
//	TestTwoFactorTOTP     for SHA1/256/512 x 6/8 digits: lib.NewTOTP seals the state
//	                      (pilinux/cryptoengine), then checkTOTPQR and checkTOTPCodes.
//	                      twofactor uses sec51/convert for time-step rounding and
//	                      big-endian encoding of the counter and the sealed fields
//	checkTOTPQR           lib.NewQR returns a square PNG (rsc.io/qr); lib.ByteToPNG
//	                      writes it
//	checkTOTPCodes        the code has the right length; ValidateTOTP accepts it,
//	                      rejects a wrong code but returns the state, and rejects
//	                      tampered sealed bytes
//	pinnedArgon2Key/Hash  outputs recorded once; any change in argon2 output fails
//	TestArgon2Hashing     lib.HashPass yields $argon2id$ that verifies only with the right
//	                      password and secret; lib.GetArgon2Key and an existing hash match
//	                      the pinned values. Covers golang.org/x/sys/cpu, which selects
//	                      x/crypto/argon2's SSE4.1 path on amd64
//	TestRateLimiter       lib.InitRateLimiter rejects "abc", "10-X" and "x-M" with ulule's
//	                      messages (pkg/errors). With "2-M" the third request from one IP
//	                      gets 429 with X-RateLimit-Limit 2 and another IP passes; with
//	                      X-Real-Ip as gin TrustedPlatform, clients behind one proxy are
//	                      limited separately

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"fmt"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/argon2"
	"github.com/pilinux/twofactor"

	glib "github.com/pilinux/gorest/lib"
	gmiddleware "github.com/pilinux/gorest/lib/middleware"
)

// TestTwoFactorTOTP covers pilinux/cryptoengine (sealing the TOTP state),
// sec51/convert (time-step rounding, big-endian encoding) and rsc.io/qr
// (enrolment QR), along the path handler/twoFA.go and service/common.go take through lib.
func TestTwoFactorTOTP(t *testing.T) {
	logIndirect(t)
	// cryptoengine persists its keys under ./keys
	t.Chdir(t.TempDir())
	if err := os.Mkdir("keys", 0o700); err != nil {
		t.Fatal(err)
	}
	const issuer = "gorest"

	for _, h := range []crypto.Hash{crypto.SHA1, crypto.SHA256, crypto.SHA512} {
		for _, digits := range []int{6, 8} {
			t.Run(fmt.Sprintf("%v/%d", h, digits), func(t *testing.T) {
				sealed, err := glib.NewTOTP("user@example.com", issuer, h, digits)
				if err != nil || len(sealed) == 0 {
					t.Fatalf("NewTOTP: %v", err)
				}
				checkTOTPQR(t, sealed, issuer)
				checkTOTPCodes(t, sealed, issuer, digits)
			})
		}
	}
}

// checkTOTPQR checks lib.NewQR returns a square PNG (rsc.io/qr) and
// lib.ByteToPNG writes it to disk.
func checkTOTPQR(t *testing.T, sealed []byte, issuer string) {
	t.Helper()
	qr, err := glib.NewQR(sealed, issuer)
	if err != nil {
		t.Fatalf("NewQR: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(qr))
	if err != nil {
		t.Fatalf("QR is not a PNG: %v", err)
	}
	if b := img.Bounds(); b.Dx() == 0 || b.Dx() != b.Dy() {
		t.Errorf("QR bounds = %v", b)
	}
	dir := t.TempDir()
	name, err := glib.ByteToPNG(qr, dir+"/")
	if err != nil {
		t.Fatalf("ByteToPNG: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
		t.Errorf("QR file not written: %v", err)
	}
}

// checkTOTPCodes checks the generated code has the right length, then that
// lib.ValidateTOTP accepts it, rejects a wrong code while still returning the
// state, and rejects tampered sealed bytes (cryptoengine).
func checkTOTPCodes(t *testing.T, sealed []byte, issuer string, digits int) {
	t.Helper()
	otp, err := twofactor.TOTPFromBytes(sealed, issuer)
	if err != nil {
		t.Fatalf("TOTPFromBytes: %v", err)
	}
	code, err := otp.OTP()
	if err != nil || len(code) != digits {
		t.Fatalf("OTP = %q, %v", code, err)
	}

	updated, err := glib.ValidateTOTP(sealed, issuer, code)
	if err != nil || len(updated) == 0 {
		t.Fatalf("ValidateTOTP(valid code): %v", err)
	}

	wrong := []byte(code)
	wrong[0] = '0' + (wrong[0]-'0'+1)%10
	state, err := glib.ValidateTOTP(sealed, issuer, string(wrong))
	if err == nil {
		t.Error("ValidateTOTP accepted a wrong code")
	}
	if len(state) == 0 {
		t.Error("ValidateTOTP must still return the sealed state on a wrong code")
	}

	tampered := bytes.Clone(sealed)
	tampered[len(tampered)/2] ^= 0xff
	if _, err := glib.ValidateTOTP(tampered, issuer, code); err == nil {
		t.Error("tampered TOTP state was accepted")
	}
}

// Pinned on one machine and verified on every other: amd64 takes the SIMD
// path selected through golang.org/x/sys/cpu, other arches the generic one.
const (
	pinnedArgon2Key  = "e1a17a9e71b9090999706952a086623fa8a5b5d5bad3558906bc6455c4988988"                                  // ggignore
	pinnedArgon2Hash = "$argon2id$v=19$m=65536,t=2,p=2$7YrCO6Wj8ploAnMo0WVuAQ$WzCWKWNqxzahgnx/LhsQMHEhJ0Ca12CiLptlttiPLGw" // ggignore
)

type argon2CompareCase struct {
	pass, secret string
	want         bool
}

// TestArgon2Hashing covers golang.org/x/sys/cpu as used by pilinux/argon2
// and x/crypto/argon2 for lib.HashPass and lib.GetArgon2Key.
func TestArgon2Hashing(t *testing.T) {
	logIndirect(t)
	cfg := glib.HashPassConfig{Memory: 64, Iterations: 2, Parallelism: 2, SaltLength: 16, KeyLength: 32}
	hash, err := glib.HashPass(cfg, "correct horse", "pepper") // ggignore
	if err != nil {
		t.Fatalf("HashPass: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("hash = %q, want argon2id encoding", hash)
	}
	for _, tc := range []argon2CompareCase{
		{"correct horse", "pepper", true}, // ggignore
		{"wrong horse", "pepper", false},  // ggignore
		{"correct horse", "salt", false},  // ggignore
	} {
		if ok, err := argon2.ComparePasswordAndHash(tc.pass, tc.secret, hash); ok != tc.want || err != nil {
			t.Errorf("Compare(%q,%q) = %v, %v; want %v", tc.pass, tc.secret, ok, err, tc.want)
		}
	}

	key := glib.GetArgon2Key([]byte("password"), []byte("somesalt-16bytes"), 32) // ggignore
	if got := hex.EncodeToString(key); got != pinnedArgon2Key {
		t.Errorf("GetArgon2Key = %s, want %s", got, pinnedArgon2Key)
	}
	if ok, err := argon2.ComparePasswordAndHash("password", "pepper", pinnedArgon2Hash); !ok || err != nil { // ggignore
		t.Errorf("pinned argon2id hash no longer verifies: %v", err)
	}
}

type rateParseCase struct{ rate, wantErr string }

// TestRateLimiter covers pkg/errors (ulule/limiter's rate parsing errors)
// along with the limiter wiring from lib.InitRateLimiter and the middleware.
func TestRateLimiter(t *testing.T) {
	logIndirect(t)
	for _, tc := range []rateParseCase{
		{"abc", "incorrect format"},
		{"10-X", "incorrect period"},
		{"x-M", "incorrect limit"},
	} {
		if _, err := glib.InitRateLimiter(tc.rate, ""); err == nil || !strings.Contains(err.Error(), tc.wantErr) {
			t.Errorf("InitRateLimiter(%q) error = %v, want %q", tc.rate, err, tc.wantErr)
		}
	}

	hit := func(r http.Handler, remote, realIP string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = remote
		if realIP != "" {
			req.Header.Set("X-Real-Ip", realIP)
		}
		r.ServeHTTP(w, req)
		return w
	}
	router := func(rate, header string) *gin.Engine {
		l, err := glib.InitRateLimiter(rate, header)
		if err != nil {
			t.Fatalf("InitRateLimiter(%q): %v", rate, err)
		}
		// same client-IP setup as the example routers
		r := gin.New()
		if err := r.SetTrustedProxies(nil); err != nil {
			t.Fatal(err)
		}
		r.TrustedPlatform = header
		r.Use(gmiddleware.RateLimit(l))
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
		return r
	}

	r := router("2-M", "")
	for i, want := range []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests} {
		if w := hit(r, "192.0.2.1:1000", ""); w.Code != want {
			t.Errorf("request %d = %d, want %d", i+1, w.Code, want)
		} else if w.Header().Get("X-RateLimit-Limit") != "2" {
			t.Errorf("X-RateLimit-Limit = %q", w.Header().Get("X-RateLimit-Limit"))
		}
	}
	if w := hit(r, "192.0.2.2:1000", ""); w.Code != http.StatusOK {
		t.Errorf("other IPv4 client = %d", w.Code)
	}

	r = router("1-M", "X-Real-Ip")
	if w := hit(r, "10.0.0.1:1", "198.51.100.1"); w.Code != http.StatusOK {
		t.Errorf("first header client = %d", w.Code)
	}
	if w := hit(r, "10.0.0.1:1", "198.51.100.2"); w.Code != http.StatusOK {
		t.Errorf("second header client behind same proxy = %d", w.Code)
	}
	if w := hit(r, "10.0.0.1:1", "198.51.100.1"); w.Code != http.StatusTooManyRequests {
		t.Errorf("repeat header client = %d", w.Code)
	}
}
