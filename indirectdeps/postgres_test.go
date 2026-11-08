package indirectdeps_test

// PostgreSQL path: database.InitDB -> gorm.io/driver/postgres -> pgx/v5 stdlib
// driver, which normalizes passwords with x/text/secure/precis.
//
// Indirect modules:
//   - github.com/jackc/pgx/v5
//   - github.com/jackc/pgpassfile
//   - github.com/jackc/pgservicefile
//   - golang.org/x/text
//
//	Function                        Purpose
//	------------------------------  -------
//	startFakePostgres               starts fakePostgres; a nil scramServer means cleartext auth
//	(*fakePostgres).handle          declines SSL/GSSENC, records startup parameters, authenticates,
//	                                sends AuthenticationOk, ParameterStatus, BackendKeyData and
//	                                ReadyForQuery, then answers each simple query (pgx's ping)
//	(*fakePostgres).authenticate    cleartext: records the password sent; SCRAM-SHA-256: SASL,
//	                                SASLContinue, SASLFinal checked by scramServer
//	pgMsg, pgAuth, pgReadTyped      message framing, Authentication messages, typed reads
//	pgFatal                         FATAL 28P01 ErrorResponse for a failed login
//	configurePostgres               RDBMS config InitDB builds its DSN from; points PGPASSFILE and
//	                                PGSERVICEFILE at missing files and empties PGSERVICE so the
//	                                developer's own files cannot change the result
//	TestPostgresSCRAMPrecis         server stores "p'a\\s \u00e9", client gets "p'a\\s\u00a0e\u0301": // ggignore
//	                                InitDB succeeds only if pgx applies precis OpaqueString before
//	                                SCRAM. User and database (space, quote) and TimeZone arrive
//	                                intact in the startup message
//	TestPostgresSCRAMWrongPassword  a wrong password fails the proof check and InitDB
//	TestPostgresPassfile            DBPASS empty: pgx reads the password from PGPASSFILE (comment
//	                                line, escaped colon) and sends "from:passfile" // ggignore
//	TestPostgresServiceFile         PGSERVICE=gorest: application_name from the service file reaches
//	                                the server, and our DSN host overrides the file's 192.0.2.1

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	gdb "github.com/pilinux/gorest/database"
)

const (
	pgProtocol3     = 196608
	pgSSLRequest    = 80877103
	pgGSSENCRequest = 80877104
)

// fakePostgres implements the startup, auth (cleartext or SCRAM-SHA-256)
// and simple-query parts of the protocol that pgx needs for InitDB.
type fakePostgres struct {
	scram *scramServer // nil means cleartext password auth

	mu        sync.Mutex
	startup   []map[string]string
	passwords []string
	authOK    int
	authBad   int
}

func startFakePostgres(t *testing.T, scram *scramServer) (*fakePostgres, string, string) {
	t.Helper()
	s := &fakePostgres{scram: scram}
	ln, host, port := listen(t)
	serve(ln, s.handle)
	return s, host, port
}

func pgMsg(typ byte, payload []byte) []byte {
	m := []byte{typ}
	m = binary.BigEndian.AppendUint32(m, uint32(len(payload)+4))
	return append(m, payload...)
}

func pgAuth(code uint32, data []byte) []byte {
	return pgMsg('R', append(binary.BigEndian.AppendUint32(nil, code), data...))
}

func pgReadTyped(r *bufio.Reader) (byte, []byte, error) {
	typ, err := r.ReadByte()
	if err != nil {
		return 0, nil, err
	}
	var n uint32
	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return 0, nil, err
	}
	p := make([]byte, n-4)
	_, err = io.ReadFull(r, p)
	return typ, p, err
}

func pgFatal(msg string) []byte {
	var p []byte
	for _, f := range [][2]string{{"S", "FATAL"}, {"V", "FATAL"}, {"C", "28P01"}, {"M", msg}} {
		p = append(p, f[0][0])
		p = append(p, f[1]+"\x00"...)
	}
	return pgMsg('E', append(p, 0))
}

func (s *fakePostgres) handle(c net.Conn) {
	r := bufio.NewReader(c)

	var params map[string]string
	for params == nil {
		var n uint32
		if err := binary.Read(r, binary.BigEndian, &n); err != nil {
			return
		}
		p := make([]byte, n-4)
		if _, err := io.ReadFull(r, p); err != nil {
			return
		}
		switch binary.BigEndian.Uint32(p) {
		case pgSSLRequest, pgGSSENCRequest:
			if _, err := c.Write([]byte{'N'}); err != nil {
				return
			}
		case pgProtocol3:
			params = map[string]string{}
			kv := bytes.Split(bytes.TrimRight(p[4:], "\x00"), []byte{0})
			for i := 0; i+1 < len(kv); i += 2 {
				params[string(kv[i])] = string(kv[i+1])
			}
		default:
			return
		}
	}
	s.mu.Lock()
	s.startup = append(s.startup, params)
	s.mu.Unlock()

	if !s.authenticate(c, r) {
		s.mu.Lock()
		s.authBad++
		s.mu.Unlock()
		_, _ = c.Write(pgFatal("password authentication failed"))
		return
	}
	s.mu.Lock()
	s.authOK++
	s.mu.Unlock()

	var out []byte
	out = append(out, pgAuth(0, nil)...)
	for _, kv := range [][2]string{{"server_version", "16.0"}, {"client_encoding", "UTF8"}, {"standard_conforming_strings", "on"}} {
		out = append(out, pgMsg('S', []byte(kv[0]+"\x00"+kv[1]+"\x00"))...)
	}
	out = append(out, pgMsg('K', []byte{0, 0, 0, 1, 0, 0, 0, 2})...)
	out = append(out, pgMsg('Z', []byte{'I'})...)
	if _, err := c.Write(out); err != nil {
		return
	}

	for {
		typ, _, err := pgReadTyped(r)
		if err != nil || typ == 'X' {
			return
		}
		// answer every simple query (pgx ping is "-- ping") with an empty result
		var resp []byte
		if typ == 'Q' {
			resp = append(resp, pgMsg('I', nil)...)
		}
		resp = append(resp, pgMsg('Z', []byte{'I'})...)
		if _, err := c.Write(resp); err != nil {
			return
		}
	}
}

func (s *fakePostgres) authenticate(c net.Conn, r *bufio.Reader) bool {
	if s.scram == nil {
		if _, err := c.Write(pgAuth(3, nil)); err != nil {
			return false
		}
		typ, p, err := pgReadTyped(r)
		if err != nil || typ != 'p' {
			return false
		}
		s.mu.Lock()
		s.passwords = append(s.passwords, string(bytes.TrimRight(p, "\x00")))
		s.mu.Unlock()
		return true
	}

	if _, err := c.Write(pgAuth(10, []byte("SCRAM-SHA-256\x00\x00"))); err != nil {
		return false
	}
	typ, p, err := pgReadTyped(r)
	if err != nil || typ != 'p' {
		return false
	}
	mech, rest, _ := bytes.Cut(p, []byte{0})
	if string(mech) != "SCRAM-SHA-256" || len(rest) < 4 {
		return false
	}
	first, err := s.scram.first(string(rest[4:]))
	if err != nil {
		return false
	}
	if _, err := c.Write(pgAuth(11, []byte(first))); err != nil {
		return false
	}
	typ, p, err = pgReadTyped(r)
	if err != nil || typ != 'p' {
		return false
	}
	final, err := s.scram.final(string(p))
	if err != nil {
		return false
	}
	_, err = c.Write(pgAuth(12, []byte(final)))
	return err == nil
}

func configurePostgres(t *testing.T, host, port, user, pass, dbName string) {
	t.Helper()
	cfg := setupConfig(t)
	rdbms := &cfg.Database.RDBMS
	rdbms.Env.Driver = "postgres"
	rdbms.Env.Host = host
	rdbms.Env.Port = port
	rdbms.Env.TimeZone = "Europe/Berlin"
	rdbms.Access.User = user
	rdbms.Access.Pass = pass
	rdbms.Access.DbName = dbName
	rdbms.Ssl.Sslmode = "disable"
	rdbms.Conn.MaxIdleConns = 1
	rdbms.Conn.MaxOpenConns = 2
	rdbms.Conn.ConnMaxLifetime = time.Minute
	rdbms.Log.LogLevel = 1
	// keep the developer's ~/.pgpass and service files out of the picture
	t.Setenv("PGPASSFILE", filepath.Join(t.TempDir(), "none"))
	t.Setenv("PGSERVICEFILE", filepath.Join(t.TempDir(), "none"))
	t.Setenv("PGSERVICE", "")
	t.Cleanup(func() { _ = gdb.CloseSQL() })
}

// TestPostgresSCRAMPrecis covers jackc/pgx/v5 (stdlib driver, DSN parsing)
// and golang.org/x/text/secure/precis: pgx normalizes the password before
// SCRAM-SHA-256, so the server only accepts the normalized form.
func TestPostgresSCRAMPrecis(t *testing.T) {
	logIndirect(t)
	// NBSP maps to a space and "e"+U+0301 composes to U+00E9 under OpaqueString
	const raw = "p'a\\s\u00a0e\u0301"  // ggignore
	const normalized = "p'a\\s \u00e9" // ggignore
	srv, host, port := startFakePostgres(t, newScramServer(sha256.New, normalized))
	configurePostgres(t, host, port, "go rest's user", raw, "gorest db")

	if db := gdb.InitDB(); db.Error != nil {
		t.Fatalf("InitDB: %v", db.Error)
	}

	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.authOK == 0 || srv.authBad != 0 {
		t.Fatalf("scram auth ok=%d bad=%d", srv.authOK, srv.authBad)
	}
	got := srv.startup[0]
	for k, want := range map[string]string{"user": "go rest's user", "database": "gorest db", "TimeZone": "Europe/Berlin"} {
		if got[k] != want {
			t.Errorf("startup %s = %q, want %q", k, got[k], want)
		}
	}
}

// TestPostgresSCRAMWrongPassword ensures a SCRAM proof mismatch fails InitDB.
func TestPostgresSCRAMWrongPassword(t *testing.T) {
	logIndirect(t)
	srv, host, port := startFakePostgres(t, newScramServer(sha256.New, "right")) // ggignore
	configurePostgres(t, host, port, "gorest", "wrong", "gorest")                // ggignore

	if db := gdb.InitDB(); db.Error == nil {
		t.Fatal("InitDB succeeded with a wrong password")
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.scram.proofBad.Load() == 0 {
		t.Error("server never rejected a proof")
	}
}

// TestPostgresPassfile covers jackc/pgpassfile: with DBPASS empty, the DSN
// carries an empty password and pgx falls back to the PGPASSFILE entry.
func TestPostgresPassfile(t *testing.T) {
	logIndirect(t)
	srv, host, port := startFakePostgres(t, nil)
	configurePostgres(t, host, port, "gorest", "", "gorest_db")

	passfile := filepath.Join(t.TempDir(), "pgpass")
	line := fmt.Sprintf("# comment\n%s:%s:gorest_db:gorest:from\\:passfile\n", host, port) // ggignore
	if err := os.WriteFile(passfile, []byte(line), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PGPASSFILE", passfile)

	if db := gdb.InitDB(); db.Error != nil {
		t.Fatalf("InitDB: %v", db.Error)
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if len(srv.passwords) == 0 || srv.passwords[0] != "from:passfile" { // ggignore
		t.Errorf("password sent = %q, want %q", srv.passwords, "from:passfile") // ggignore
	}
}

// TestPostgresServiceFile covers jackc/pgservicefile: PGSERVICE settings are
// merged into our keyword/value DSN, and DSN values take precedence.
func TestPostgresServiceFile(t *testing.T) {
	logIndirect(t)
	srv, host, port := startFakePostgres(t, nil)
	configurePostgres(t, host, port, "gorest", "secret", "gorest_db") // ggignore

	svc := filepath.Join(t.TempDir(), "pg_service.conf")
	conf := "# comment\n[gorest]\nhost=192.0.2.1\nconnect_timeout=2\napplication_name=gorest-indirectdeps\n"
	if err := os.WriteFile(svc, []byte(conf), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PGSERVICEFILE", svc)
	t.Setenv("PGSERVICE", "gorest")

	if db := gdb.InitDB(); db.Error != nil {
		t.Fatalf("InitDB: %v", db.Error)
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if got := srv.startup[0]["application_name"]; got != "gorest-indirectdeps" {
		t.Errorf("application_name from service file = %q", got)
	}
}
