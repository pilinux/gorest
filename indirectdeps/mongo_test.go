package indirectdeps_test

// MongoDB path: database.InitMongo -> mongo-driver/v2 with compressors and
// SCRAM credentials from MONGO_URI; stringprep normalizes with
// x/text/unicode/norm.
//
// Indirect modules:
//   - github.com/klauspost/compress
//   - github.com/xdg-go/scram
//   - github.com/xdg-go/stringprep
//   - golang.org/x/text
//
//	Function                  Purpose
//	------------------------  -------
//	opReply ... opMsg         wire opcodes the fake reads and writes
//	newFakeMongo              fake with an optional scramServer and the compressors it agrees to
//	e, lookup                 bson.E shorthand; finds a key in a bson.D
//	(*fakeMongo).handle       unwraps OP_COMPRESSED (counted per compressor), decodes OP_MSG
//	                          or legacy OP_QUERY, replies with the same opcode and compressor
//	(*fakeMongo).reply        hello/isMaster: agrees to supported compressors, advertises
//	                          SCRAM-SHA-1/256; saslStart/saslContinue: runs scramServer and
//	                          records the username; other commands: ok 1
//	configureMongo            sets MONGO_URI, app name, pool size, TTL; closes the client
//	startFakeMongo            serves the fake on loopback, returns host:port
//	TestMongoWireCompression  compressors=snappy, then zstd: InitMongo succeeds, the fake
//	                          receives messages compressed with that algorithm, and ping
//	                          arrives
//	TestMongoSCRAM            SCRAM-SHA-256: server stores "passIX", client gets "pa\u00adss\u2168"; // ggignore
//	                          InitMongo succeeds only if SASLprep drops the soft hyphen and NFKC
//	                          maps U+2168; a wrong password fails at the proof check. SCRAM-SHA-1: server
//	                          stores hex(md5(user:mongo:pass)), the driver's SHA-1 digest.
//	                          Each success sends the user as "go=2Crest=3Duser" (RFC 5802
//	                          escaping by xdg-go/scram)
//
// The fake compresses with the driver's own codec (which wraps
// klauspost/compress), so TestMongoWireCompression proves negotiation and a
// working round trip, not byte compatibility with a real mongod.

import (
	"bufio"
	"bytes"
	"crypto/md5"  // #nosec G501 -- MongoDB SCRAM-SHA-1 password digest
	"crypto/sha1" // #nosec G505 -- MongoDB SCRAM-SHA-1
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/wiremessage"

	gdb "github.com/pilinux/gorest/database"
)

const (
	opReply      = 1
	opQuery      = 2004
	opCompressed = 2012
	opMsg        = 2013
)

// fakeMongo answers hello, SASL and generic commands over the wire
// protocol, including OP_COMPRESSED.
type fakeMongo struct {
	scram       *scramServer
	compressors []string

	mu         sync.Mutex
	commands   []string
	compressed map[wiremessage.CompressorID]int
	authOK     int
	authBad    int
	scramUser  string
}

func newFakeMongo(scram *scramServer, compressors ...string) *fakeMongo {
	return &fakeMongo{scram: scram, compressors: compressors, compressed: map[wiremessage.CompressorID]int{}}
}

func e(k string, v any) bson.E { return bson.E{Key: k, Value: v} }

func lookup(d bson.D, key string) (any, bool) {
	for _, el := range d {
		if el.Key == key {
			return el.Value, true
		}
	}
	return nil, false
}

func (s *fakeMongo) handle(c net.Conn) {
	r := bufio.NewReader(c)
	for {
		var hdr [16]byte
		if _, err := io.ReadFull(r, hdr[:]); err != nil {
			return
		}
		size := int32(binary.LittleEndian.Uint32(hdr[0:]))
		reqID := int32(binary.LittleEndian.Uint32(hdr[4:]))
		op := int32(binary.LittleEndian.Uint32(hdr[12:]))
		body := make([]byte, size-16)
		if _, err := io.ReadFull(r, body); err != nil {
			return
		}

		comp := wiremessage.CompressorNoOp
		if op == opCompressed {
			op = int32(binary.LittleEndian.Uint32(body[0:]))
			comp = wiremessage.CompressorID(body[8])
			var err error
			body, err = driver.DecompressPayload(body[9:], driver.CompressionOpts{
				Compressor:       comp,
				UncompressedSize: int32(binary.LittleEndian.Uint32(body[4:])),
			})
			if err != nil {
				return
			}
			s.mu.Lock()
			s.compressed[comp]++
			s.mu.Unlock()
		}

		var cmd bson.D
		switch op {
		case opMsg:
			// flagBits(4), then a kind-0 section holding the command document
			if len(body) < 5 || body[4] != 0 {
				return
			}
			if err := bson.Unmarshal(body[5:], &cmd); err != nil {
				return
			}
		case opQuery:
			rest := body[4:]
			_, rest, _ = bytes.Cut(rest, []byte{0})
			if err := bson.Unmarshal(rest[8:], &cmd); err != nil {
				return
			}
		default:
			return
		}

		reply, err := bson.Marshal(s.reply(cmd))
		if err != nil {
			return
		}
		var out []byte
		if op == opMsg {
			out = append(make([]byte, 4), 0)
			out = append(out, reply...)
		} else {
			out = binary.LittleEndian.AppendUint32(nil, 0)
			out = binary.LittleEndian.AppendUint64(out, 0)
			out = binary.LittleEndian.AppendUint32(out, 0)
			out = binary.LittleEndian.AppendUint32(out, 1)
			out = append(out, reply...)
			op = opReply
		}
		if comp != wiremessage.CompressorNoOp {
			packed, err := driver.CompressPayload(out, driver.CompressionOpts{Compressor: comp, ZlibLevel: 6, ZstdLevel: 6})
			if err != nil {
				return
			}
			inner := binary.LittleEndian.AppendUint32(nil, uint32(op))
			inner = binary.LittleEndian.AppendUint32(inner, uint32(len(out)))
			inner = append(inner, byte(comp))
			out = append(inner, packed...)
			op = opCompressed
		}
		msg := binary.LittleEndian.AppendUint32(nil, uint32(16+len(out)))
		msg = binary.LittleEndian.AppendUint32(msg, uint32(reqID)+1000)
		msg = binary.LittleEndian.AppendUint32(msg, uint32(reqID))
		msg = binary.LittleEndian.AppendUint32(msg, uint32(op))
		if _, err := c.Write(append(msg, out...)); err != nil {
			return
		}
	}
}

func (s *fakeMongo) reply(cmd bson.D) bson.D {
	if len(cmd) == 0 {
		return bson.D{e("ok", 0.0)}
	}
	name := strings.ToLower(cmd[0].Key)
	s.mu.Lock()
	s.commands = append(s.commands, name)
	s.mu.Unlock()

	switch name {
	case "hello", "ismaster":
		d := bson.D{
			e("helloOk", true), e("isWritablePrimary", true), e("ismaster", true),
			e("maxBsonObjectSize", int32(16*1024*1024)), e("maxMessageSizeBytes", int32(48000000)),
			e("maxWriteBatchSize", int32(100000)), e("localTime", bson.NewDateTimeFromTime(time.Now())),
			e("logicalSessionTimeoutMinutes", int32(30)), e("connectionId", int32(1)),
			e("minWireVersion", int32(0)), e("maxWireVersion", int32(21)), e("readOnly", false),
		}
		if v, ok := lookup(cmd, "compression"); ok {
			var agreed bson.A
			for _, c := range v.(bson.A) {
				if slices.Contains(s.compressors, c.(string)) {
					agreed = append(agreed, c)
				}
			}
			d = append(d, e("compression", agreed))
		}
		if _, ok := lookup(cmd, "saslSupportedMechs"); ok {
			d = append(d, e("saslSupportedMechs", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}))
		}
		return append(d, e("ok", 1.0))

	case "saslstart", "saslcontinue":
		v, _ := lookup(cmd, "payload")
		payload, _ := v.(bson.Binary)
		var (
			out  string
			err  error
			done = name == "saslcontinue"
		)
		if s.scram == nil {
			err = fmt.Errorf("auth not configured")
		} else if done {
			out, err = s.scram.final(string(payload.Data))
		} else {
			out, err = s.scram.first(string(payload.Data))
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		if err != nil {
			s.authBad++
			return bson.D{e("ok", 0.0), e("errmsg", "Authentication failed."), e("code", int32(18)), e("codeName", "AuthenticationFailed")}
		}
		if done {
			s.authOK++
			s.scramUser = s.scram.user
		}
		return bson.D{e("conversationId", int32(1)), e("done", done), e("payload", bson.Binary{Data: []byte(out)}), e("ok", 1.0)}
	}
	return bson.D{e("ok", 1.0)}
}

func configureMongo(t *testing.T, uri string) {
	t.Helper()
	cfg := setupConfig(t)
	cfg.Database.MongoDB.Env.URI = uri
	cfg.Database.MongoDB.Env.AppName = "gorest-indirectdeps"
	cfg.Database.MongoDB.Env.PoolSize = 5
	cfg.Database.MongoDB.Env.ConnTTL = 5
	t.Cleanup(func() { _ = gdb.CloseMongo() })
}

func startFakeMongo(t *testing.T, s *fakeMongo) string {
	t.Helper()
	ln, _, _ := listen(t)
	serve(ln, s.handle)
	return ln.Addr().String()
}

type mongoCompressionCase struct {
	name string
	id   wiremessage.CompressorID
}

// TestMongoWireCompression covers klauspost/compress (snappy, zstd): with
// compressors in MONGO_URI the driver compresses commands after hello.
func TestMongoWireCompression(t *testing.T) {
	logIndirect(t)
	for _, tc := range []mongoCompressionCase{
		{"snappy", wiremessage.CompressorSnappy},
		{"zstd", wiremessage.CompressorZstd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := newFakeMongo(nil, tc.name)
			addr := startFakeMongo(t, srv)
			configureMongo(t, "mongodb://"+addr+"/?directConnection=true&compressors="+tc.name)

			if _, err := gdb.InitMongo(); err != nil {
				t.Fatalf("InitMongo: %v", err)
			}
			srv.mu.Lock()
			defer srv.mu.Unlock()
			if srv.compressed[tc.id] == 0 {
				t.Errorf("no %s-compressed messages; commands: %q", tc.name, srv.commands)
			}
			if !slices.Contains(srv.commands, "ping") {
				t.Errorf("ping not received; commands: %q", srv.commands)
			}
		})
	}
}

type mongoSCRAMCase struct {
	mech      string
	server    *scramServer
	pass      string
	wantError bool
}

// TestMongoSCRAM covers xdg-go/scram and xdg-go/stringprep (with
// golang.org/x/text/unicode/norm): credentials in MONGO_URI are
// authenticated with SCRAM, SASLprepping the password for SHA-256.
func TestMongoSCRAM(t *testing.T) {
	logIndirect(t)
	const user = "go,rest=user"
	// SASLprep drops the soft hyphen and NFKC-maps U+2168 to "IX"
	const rawPass = "pa\u00adss\u2168"                        // ggignore
	const preppedPass = "passIX"                              // ggignore
	sha1Digest := md5.Sum([]byte(user + ":mongo:" + rawPass)) // #nosec G401 -- protocol requirement

	for _, tc := range []mongoSCRAMCase{
		{"SCRAM-SHA-256", newScramServer(sha256.New, preppedPass), rawPass, false},
		{"SCRAM-SHA-256", newScramServer(sha256.New, preppedPass), "wrong", true}, // ggignore
		{"SCRAM-SHA-1", newScramServer(sha1.New, hex.EncodeToString(sha1Digest[:])), rawPass, false},
	} {
		t.Run(fmt.Sprintf("%s/wantError=%v", tc.mech, tc.wantError), func(t *testing.T) {
			srv := newFakeMongo(tc.server)
			addr := startFakeMongo(t, srv)
			uri := fmt.Sprintf("mongodb://%s:%s@%s/?directConnection=true&authSource=admin&authMechanism=%s",
				url.QueryEscape(user), url.QueryEscape(tc.pass), addr, tc.mech)
			configureMongo(t, uri)

			_, err := gdb.InitMongo()
			if tc.wantError {
				if err == nil {
					t.Fatal("InitMongo succeeded with a wrong password")
				}
				// fails for the proof, not for an earlier error
				if tc.server.proofBad.Load() == 0 {
					t.Errorf("server never rejected a SCRAM proof: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("InitMongo: %v", err)
			}
			srv.mu.Lock()
			defer srv.mu.Unlock()
			if srv.authOK == 0 {
				t.Fatal("no successful SCRAM conversation")
			}
			if want := "go=2Crest=3Duser"; srv.scramUser != want {
				t.Errorf("SCRAM username = %q, want %q", srv.scramUser, want)
			}
		})
	}
}
