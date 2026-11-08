package indirectdeps_test

// Shared fixtures. None imports an indirect module; they only set up the
// conditions under which gorest code reaches one.
//
// Indirect modules:
//   - none
//
//	Identifier            Purpose
//	--------------------  -------
//	repoRoot              repo root, resolved before any t.Chdir; indirect_test.go reads
//	                      go.mod and indirectdeps/*_test.go from it
//	setupConfig           chdir to a temp dir, write a .env, run gconfig.Config() and return
//	                      GetConfig(): InitDB, InitRedis and InitMongo read the fields each
//	                      test sets, and files written to cwd stay out of the repo
//	listen                127.0.0.1:0 listener for a fake server, closed at test cleanup
//	serve                 accept loop, one goroutine per connection running the fake's handler
//	scramServer           server side of SCRAM (RFC 5802/7677) for the fake Postgres and
//	                      MongoDB servers, built on stdlib crypto only
//	newScramServer        stores the password as the server keeps it (already normalized),
//	                      a random 16-byte salt and 4096 iterations
//	(*scramServer).first  parses client-first (GS2 header n,,, user, nonce); returns
//	                      server-first with the combined nonce, salt and iterations
//	(*scramServer).final  checks the nonce, recomputes ClientKey from the stored password,
//	                      rejects a non-matching proof, returns the server signature
//	(*scramServer).hmac   HMAC with the mechanism's hash (SHA-1 or SHA-256)
//	linkedModules         modules and versions compiled into this test binary
//	                      (debug.ReadBuildInfo), so tag, GOARCH and Go version count
//	logIndirect           prints the indirectUses modules of the running test that
//	                      are linked, with their versions; shown with go test -v
//
// scramServer shares no code with pgx or xdg-go/scram, so a bump that changes
// how the client derives the proof or normalizes the password fails the SCRAM
// tests instead of passing on both sides.

import (
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"net"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"testing"

	gconfig "github.com/pilinux/gorest/config"
)

// repoRoot is resolved before any test changes the working directory.
var repoRoot = func() string {
	p, err := filepath.Abs("..")
	if err != nil {
		panic(err)
	}
	return p
}()

// setupConfig loads a minimal config from a throwaway working directory.
// Libraries that write to cwd (cryptoengine keys/, .env) stay out of the repo.
func setupConfig(t *testing.T) *gconfig.Configuration {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".env", []byte("# indirectdeps test\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	if err := gconfig.Config(); err != nil {
		t.Fatalf("config.Config: %v", err)
	}
	cfg := gconfig.GetConfig()
	if cfg == nil {
		t.Fatal("config.GetConfig returned nil")
	}
	return cfg
}

// listen opens a loopback listener closed at test cleanup.
func listen(t *testing.T) (net.Listener, string, string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	host, port, _ := net.SplitHostPort(ln.Addr().String())
	return ln, host, port
}

// serve accepts connections until the listener closes.
func serve(ln net.Listener, handle func(net.Conn)) {
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer func() { _ = c.Close() }()
				handle(c)
			}()
		}
	}()
}

// scramServer is the server side of RFC 5802/7677 SCRAM, used by the fake
// Postgres and MongoDB servers to check the client's proof.
type scramServer struct {
	newHash  func() hash.Hash
	password string // already normalized the way the server stores it
	salt     []byte
	iter     int

	clientFirstBare string
	serverFirst     string
	nonce           string
	user            string
}

func newScramServer(newHash func() hash.Hash, password string) *scramServer {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	return &scramServer{newHash: newHash, password: password, salt: salt, iter: 4096}
}

func (s *scramServer) first(clientFirst string) (string, error) {
	// strip the GS2 header "n,," (no channel binding)
	parts := strings.SplitN(clientFirst, ",", 3)
	if len(parts) != 3 || parts[0] != "n" {
		return "", fmt.Errorf("unexpected gs2 header in %q", clientFirst)
	}
	s.clientFirstBare = parts[2]
	var cnonce string
	for attr := range strings.SplitSeq(s.clientFirstBare, ",") {
		if v, ok := strings.CutPrefix(attr, "n="); ok {
			s.user = v
		}
		if v, ok := strings.CutPrefix(attr, "r="); ok {
			cnonce = v
		}
	}
	if cnonce == "" {
		return "", errors.New("client nonce missing")
	}
	snonce := make([]byte, 18)
	_, _ = rand.Read(snonce)
	s.nonce = cnonce + base64.StdEncoding.EncodeToString(snonce)
	s.serverFirst = fmt.Sprintf("r=%s,s=%s,i=%d", s.nonce, base64.StdEncoding.EncodeToString(s.salt), s.iter)
	return s.serverFirst, nil
}

func (s *scramServer) final(clientFinal string) (string, error) {
	withoutProof, proofB64, ok := strings.Cut(clientFinal, ",p=")
	if !ok {
		return "", errors.New("client proof missing")
	}
	if !strings.Contains(withoutProof, "r="+s.nonce) {
		return "", errors.New("nonce mismatch")
	}
	proof, err := base64.StdEncoding.DecodeString(proofB64)
	if err != nil {
		return "", err
	}

	h := s.newHash()
	salted, err := pbkdf2.Key(s.newHash, s.password, s.salt, s.iter, h.Size())
	if err != nil {
		return "", err
	}
	clientKey := s.hmac(salted, "Client Key")
	h.Write(clientKey)
	storedKey := h.Sum(nil)
	authMsg := s.clientFirstBare + "," + s.serverFirst + "," + withoutProof
	clientSig := s.hmac(storedKey, authMsg)
	if len(proof) != len(clientSig) {
		return "", errors.New("proof length mismatch")
	}
	recovered := make([]byte, len(proof))
	subtle.XORBytes(recovered, proof, clientSig)
	if !hmac.Equal(recovered, clientKey) {
		return "", errors.New("authentication failed")
	}
	serverSig := s.hmac(s.hmac(salted, "Server Key"), authMsg)
	return "v=" + base64.StdEncoding.EncodeToString(serverSig), nil
}

func (s *scramServer) hmac(key []byte, msg string) []byte {
	m := hmac.New(s.newHash, key)
	m.Write([]byte(msg))
	return m.Sum(nil)
}

// linkedModules maps each module compiled into this test binary to its version,
// so the result follows the active build tags, GOARCH and Go version.
var linkedModules = sync.OnceValue(func() map[string]string {
	mods := map[string]string{}
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return mods
	}
	for _, d := range bi.Deps {
		if d.Replace != nil {
			d = d.Replace
		}
		mods[d.Path] = d.Version
	}
	return mods
})

// logIndirect prints the indirect modules the calling test exercises (from
// indirectUses) that are linked into this binary, with their linked versions.
func logIndirect(t *testing.T) {
	t.Helper()
	linked := linkedModules()
	if len(linked) == 0 {
		fmt.Printf("[indirectdeps] %s: no build info in test binary\n", t.Name())
		return
	}
	var mods []string
	for m, use := range indirectUses {
		if _, ok := linked[m]; ok && slices.Contains(use.tests, t.Name()) {
			mods = append(mods, m)
		}
	}
	slices.Sort(mods)

	var b strings.Builder
	fmt.Fprintf(&b, "[indirectdeps] %s uses:\n", t.Name())
	if len(mods) == 0 {
		b.WriteString("  - no indirect module\n")
	}
	for _, m := range mods {
		fmt.Fprintf(&b, "  - %s %s\n", m, linked[m])
	}
	fmt.Print(b.String())
}
