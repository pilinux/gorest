package indirectdeps_test

// MySQL/MariaDB path: database.InitDB -> gorm.io/driver/mysql -> go-sql-
// driver/mysql, which signs MariaDB's client_ed25519 challenge with
// edwards25519.
//
// Indirect modules:
//   - filippo.io/edwards25519
//
//	Function                           Purpose
//	---------------------------------  -------
//	startFakeMySQL                     starts fakeMySQL; the 32-byte password is the ed25519 seed, so
//	                                   stdlib crypto/ed25519 derives the key MariaDB derives
//	(*fakeMySQL).handle                handshake, auth switch to client_ed25519 with a 32-byte scramble,
//	                                   then COM_QUERY/COM_PING; records user, db, queries, auth result
//	(*fakeMySQL).writeVersion          result set for SELECT VERSION(), which gorm's mysql dialector
//	                                   runs in Initialize
//	myReadPacket, myWritePacket        packet framing (3-byte length, sequence id)
//	myLenEncStr, myErr                 length-encoded string, ERR packet
//	nonZeroRandom                      scramble bytes without NUL, which the driver treats as a terminator
//	configureMySQL                     RDBMS config InitDB builds its DSN from (driver mysql, TLS off)
//	TestMySQLEd25519Auth               InitDB succeeds only if the driver's signature equals
//	                                   ed25519.Sign(seed, scramble); user, db and the DSN's
//	                                   SET NAMES utf8mb4 reach the server
//	TestMySQLEd25519AuthWrongPassword  a signature from another password is rejected and InitDB fails

import (
	"bufio"
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"io"
	"net"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	gdatabase "github.com/pilinux/gorest/database"
)

// MySQL capability flags used by the fake server.
const (
	myLongPassword   = 0x1
	myLongFlag       = 0x4
	myConnectWithDB  = 0x8
	myProtocol41     = 0x200
	myTransactions   = 0x2000
	mySecureConn     = 0x8000
	myMultiResults   = 0x20000
	myPluginAuth     = 0x80000
	myServerVersion  = "10.11.6-MariaDB"
	myEd25519Plugin  = "client_ed25519"
	myNativePassword = "mysql_native_password"
)

// fakeMySQL speaks enough of the MySQL protocol for database.InitDB:
// handshake, auth switch to MariaDB client_ed25519, SELECT VERSION(),
// SET NAMES and COM_PING.
type fakeMySQL struct {
	priv ed25519.PrivateKey // derived from the 32-byte password, like MariaDB

	mu      sync.Mutex
	users   []string
	dbs     []string
	queries []string
	authOK  int
	authBad int
}

func startFakeMySQL(t *testing.T, password string) (*fakeMySQL, string, string) {
	t.Helper()
	if len(password) != ed25519.SeedSize {
		t.Fatalf("password must be %d bytes so stdlib ed25519 can derive the key", ed25519.SeedSize)
	}
	s := &fakeMySQL{priv: ed25519.NewKeyFromSeed([]byte(password))}
	ln, host, port := listen(t)
	serve(ln, s.handle)
	return s, host, port
}

func myReadPacket(r *bufio.Reader) (byte, []byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return 0, nil, err
	}
	n := int(hdr[0]) | int(hdr[1])<<8 | int(hdr[2])<<16
	p := make([]byte, n)
	_, err := io.ReadFull(r, p)
	return hdr[3], p, err
}

func myWritePacket(w io.Writer, seq byte, p []byte) error {
	hdr := []byte{byte(len(p)), byte(len(p) >> 8), byte(len(p) >> 16), seq}
	_, err := w.Write(append(hdr, p...))
	return err
}

func myLenEncStr(s string) []byte {
	return append([]byte{byte(len(s))}, s...)
}

// nonZeroRandom avoids NUL bytes, which the driver treats as terminators.
func nonZeroRandom(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		if b[i] == 0 {
			b[i] = 1
		}
	}
	return b
}

var (
	myOK  = []byte{0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00}
	myEOF = []byte{0xfe, 0x00, 0x00, 0x02, 0x00}
)

func myErr(code uint16, state, msg string) []byte {
	p := []byte{0xff}
	p = binary.LittleEndian.AppendUint16(p, code)
	p = append(p, '#')
	p = append(p, state...)
	return append(p, msg...)
}

func (s *fakeMySQL) handle(c net.Conn) {
	r := bufio.NewReader(c)
	caps := uint32(myLongPassword | myLongFlag | myConnectWithDB | myProtocol41 |
		myTransactions | mySecureConn | myMultiResults | myPluginAuth)
	scramble := nonZeroRandom(20)

	hs := []byte{10}
	hs = append(hs, myServerVersion+"\x00"...)
	hs = append(hs, 1, 0, 0, 0)
	hs = append(hs, scramble[:8]...)
	hs = append(hs, 0)
	hs = binary.LittleEndian.AppendUint16(hs, uint16(caps))
	hs = append(hs, 45)
	hs = binary.LittleEndian.AppendUint16(hs, 2)
	hs = binary.LittleEndian.AppendUint16(hs, uint16(caps>>16))
	hs = append(hs, 21)
	hs = append(hs, make([]byte, 10)...)
	hs = append(hs, scramble[8:]...)
	hs = append(hs, 0)
	hs = append(hs, myNativePassword+"\x00"...)
	if myWritePacket(c, 0, hs) != nil {
		return
	}

	// handshake response: caps(4) maxPacket(4) charset(1) filler(23) user\0 auth db\0 plugin\0
	seq, resp, err := myReadPacket(r)
	if err != nil || len(resp) < 33 {
		return
	}
	clientCaps := binary.LittleEndian.Uint32(resp)
	rest := resp[32:]
	user, rest, _ := bytes.Cut(rest, []byte{0})
	if len(rest) == 0 {
		return
	}
	rest = rest[1+int(rest[0]):]
	var db []byte
	if clientCaps&myConnectWithDB != 0 {
		db, _, _ = bytes.Cut(rest, []byte{0})
	}
	s.mu.Lock()
	s.users = append(s.users, string(user))
	s.dbs = append(s.dbs, string(db))
	s.mu.Unlock()

	// switch to MariaDB ed25519 auth with a 32-byte scramble
	edScramble := nonZeroRandom(32)
	sw := append([]byte{0xfe}, myEd25519Plugin+"\x00"...)
	sw = append(sw, edScramble...)
	sw = append(sw, 0)
	if myWritePacket(c, seq+1, sw) != nil {
		return
	}
	seq, sig, err := myReadPacket(r)
	if err != nil {
		return
	}
	if !bytes.Equal(sig, ed25519.Sign(s.priv, edScramble)) {
		s.mu.Lock()
		s.authBad++
		s.mu.Unlock()
		_ = myWritePacket(c, seq+1, myErr(1045, "28000", "Access denied"))
		return
	}
	s.mu.Lock()
	s.authOK++
	s.mu.Unlock()
	if myWritePacket(c, seq+1, myOK) != nil {
		return
	}

	for {
		_, cmd, err := myReadPacket(r)
		if err != nil || len(cmd) == 0 {
			return
		}
		switch cmd[0] {
		case 0x01: // COM_QUIT
			return
		case 0x03: // COM_QUERY
			q := string(cmd[1:])
			s.mu.Lock()
			s.queries = append(s.queries, q)
			s.mu.Unlock()
			if strings.EqualFold(strings.TrimSpace(q), "SELECT VERSION()") {
				if s.writeVersion(c) != nil {
					return
				}
				continue
			}
			if myWritePacket(c, 1, myOK) != nil {
				return
			}
		default: // COM_PING and anything else
			if myWritePacket(c, 1, myOK) != nil {
				return
			}
		}
	}
}

func (s *fakeMySQL) writeVersion(w io.Writer) error {
	col := slices.Concat(myLenEncStr("def"), myLenEncStr(""), myLenEncStr(""), myLenEncStr(""),
		myLenEncStr("VERSION()"), myLenEncStr(""),
		[]byte{0x0c, 33, 0, 0, 1, 0, 0, 0xfd, 0, 0, 0x1f, 0, 0})
	for i, p := range [][]byte{{1}, col, myEOF, myLenEncStr(myServerVersion), myEOF} {
		if err := myWritePacket(w, byte(i+1), p); err != nil {
			return err
		}
	}
	return nil
}

func configureMySQL(t *testing.T, host, port, user, pass, dbName string) {
	t.Helper()
	cfg := setupConfig(t)
	rdbms := &cfg.Database.RDBMS
	rdbms.Env.Driver = "mysql"
	rdbms.Env.Host = host
	rdbms.Env.Port = port
	rdbms.Access.User = user
	rdbms.Access.Pass = pass
	rdbms.Access.DbName = dbName
	rdbms.Ssl.Sslmode = "disable"
	rdbms.Conn.MaxIdleConns = 1
	rdbms.Conn.MaxOpenConns = 2
	rdbms.Conn.ConnMaxLifetime = time.Minute
	rdbms.Log.LogLevel = 1
	t.Cleanup(func() { _ = gdatabase.CloseSQL() })
}

// TestMySQLEd25519Auth covers filippo.io/edwards25519: go-sql-driver/mysql
// signs MariaDB's client_ed25519 challenge with it when database.InitDB
// connects via the DSN built from config.
func TestMySQLEd25519Auth(t *testing.T) {
	logIndirect(t)
	// 32 bytes so stdlib ed25519 derives the same key MariaDB does
	const password = "p@ss:w/rd?&=#%;'\" with spaces!!!"
	srv, host, port := startFakeMySQL(t, password)
	configureMySQL(t, host, port, "gorest_user", password, "gorest_db")

	db := gdatabase.InitDB()
	if db.Error != nil {
		t.Fatalf("InitDB: %v", db.Error)
	}

	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.authOK == 0 || srv.authBad != 0 {
		t.Fatalf("ed25519 auth: ok=%d bad=%d, want ok>0 bad=0", srv.authOK, srv.authBad)
	}
	if srv.users[0] != "gorest_user" || srv.dbs[0] != "gorest_db" {
		t.Errorf("handshake user/db = %q/%q", srv.users[0], srv.dbs[0])
	}
	if !slices.ContainsFunc(srv.queries, func(q string) bool { return strings.HasPrefix(q, "SET NAMES utf8mb4") }) {
		t.Errorf("charset=utf8mb4 from the DSN was not applied; queries: %q", srv.queries)
	}
}

// TestMySQLEd25519AuthWrongPassword ensures a bad ed25519 signature fails InitDB.
func TestMySQLEd25519AuthWrongPassword(t *testing.T) {
	logIndirect(t)
	srv, host, port := startFakeMySQL(t, strings.Repeat("a", ed25519.SeedSize))
	configureMySQL(t, host, port, "gorest_user", strings.Repeat("b", ed25519.SeedSize), "gorest_db")

	if db := gdatabase.InitDB(); db.Error == nil {
		t.Fatal("InitDB succeeded with a wrong password")
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.authBad == 0 {
		t.Error("server never saw an ed25519 signature")
	}
}
