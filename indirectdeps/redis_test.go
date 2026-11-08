package indirectdeps_test

// Redis path: database.InitRedis -> radix/v4 pool, which uses clock for init
// timing, timeouts and the periodic PING.
//
// Indirect modules:
//   - github.com/tilinna/clock
//
//	Function                    Purpose
//	--------------------------  -------
//	startFakeRedis              starts fakeRedis on loopback
//	respReadCommand             reads one RESP2 command (array of bulk strings or inline)
//	respBulk, respInt, respNil  RESP2 bulk string, integer and nil replies
//	(*fakeRedis).handle         runs exec for each command until the client disconnects
//	(*fakeRedis).exec           runs the redisCmds handler; unknown commands and wrong argument
//	                            counts are recorded as unknown
//	redisCmds                   PING (counted), SET, GET, GETDEL, EXISTS, DEL, HSET, HGET,
//	                            HINCRBY, HGETALL, HDEL: one fakeRedis method each
//	(*fakeRedis).count          EXISTS/DEL key count across kv and hashes
//	(*fakeRedis).hash           creates a hash on first write
//	TestRedisPoolAndCommands    InitRedis with PoolSize 10, then in order the subtests below
//	checkRedisBlacklist         SET EXAT (handler/logout.go), EXISTS (service.IsTokenAllowed)
//	checkRedisCounter           HINCRBY (handler/passwordReset.go, service/twoFALimiter.go)
//	checkRedisHash              HSET, HGETALL (example apps), HGET, DEL (passwordReset.go,
//	                            twoFALimiter.go)
//	checkRedisOneShot           SET, GETDEL twice (handler/verification.go); second is nil
//	waitForRedisPing            within 3s the pool sends a PING (default interval 5s /
//	                            PoolSize), and no unknown command was seen

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mediocregopher/radix/v4"

	gconfig "github.com/pilinux/gorest/config"
	gdatabase "github.com/pilinux/gorest/database"
	gservice "github.com/pilinux/gorest/service"
)

// fakeRedis is a RESP2 server covering the commands gorest issues.
type fakeRedis struct {
	mu      sync.Mutex
	kv      map[string]string
	hashes  map[string]map[string]string
	pings   int
	unknown []string
}

func startFakeRedis(t *testing.T) (*fakeRedis, string, string) {
	t.Helper()
	s := &fakeRedis{kv: map[string]string{}, hashes: map[string]map[string]string{}}
	ln, host, port := listen(t)
	serve(ln, s.handle)
	return s, host, port
}

func respReadCommand(r *bufio.Reader) ([]string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(line, "*") {
		return strings.Fields(line), nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(line[1:]))
	if err != nil {
		return nil, err
	}
	args := make([]string, n)
	for i := range args {
		hdr, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(strings.TrimSpace(hdr[1:]))
		if err != nil {
			return nil, err
		}
		buf := make([]byte, size+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		args[i] = string(buf[:size])
	}
	return args, nil
}

func respBulk(s string) string { return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s) }
func respInt(n int) string     { return fmt.Sprintf(":%d\r\n", n) }

const respNil = "$-1\r\n"

func (s *fakeRedis) handle(c net.Conn) {
	r := bufio.NewReader(c)
	for {
		args, err := respReadCommand(r)
		if err != nil || len(args) == 0 {
			return
		}
		if _, err := io.WriteString(c, s.exec(args)); err != nil {
			return
		}
	}
}

func (s *fakeRedis) exec(args []string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if fn, ok := redisCmds[strings.ToUpper(args[0])]; ok {
		if reply, ok := fn(s, args); ok {
			return reply
		}
	}
	s.unknown = append(s.unknown, strings.Join(args, " "))
	return "-ERR unknown command\r\n"
}

// redisCmds maps each supported command to its handler; a handler returns
// false on a wrong argument count, which exec records as unknown.
var redisCmds = map[string]func(*fakeRedis, []string) (string, bool){
	"PING":    (*fakeRedis).ping,
	"SET":     (*fakeRedis).set,
	"GET":     (*fakeRedis).get,
	"GETDEL":  (*fakeRedis).getdel,
	"EXISTS":  (*fakeRedis).exists,
	"DEL":     (*fakeRedis).del,
	"HSET":    (*fakeRedis).hset,
	"HGET":    (*fakeRedis).hget,
	"HINCRBY": (*fakeRedis).hincrby,
	"HGETALL": (*fakeRedis).hgetall,
	"HDEL":    (*fakeRedis).hdel,
}

func (s *fakeRedis) ping([]string) (string, bool) {
	s.pings++
	return "+PONG\r\n", true
}

func (s *fakeRedis) set(args []string) (string, bool) {
	if len(args) < 3 {
		return "", false
	}
	s.kv[args[1]] = args[2]
	return "+OK\r\n", true
}

func (s *fakeRedis) get(args []string) (string, bool) {
	if len(args) != 2 {
		return "", false
	}
	v, ok := s.kv[args[1]]
	if !ok {
		return respNil, true
	}
	return respBulk(v), true
}

func (s *fakeRedis) getdel(args []string) (string, bool) {
	reply, ok := s.get(args)
	if ok {
		delete(s.kv, args[1])
	}
	return reply, ok
}

// count returns how many keys exist in kv or hashes, deleting them if del.
func (s *fakeRedis) count(keys []string, del bool) int {
	n := 0
	for _, k := range keys {
		_, inKV := s.kv[k]
		_, inHash := s.hashes[k]
		if !inKV && !inHash {
			continue
		}
		n++
		if del {
			delete(s.kv, k)
			delete(s.hashes, k)
		}
	}
	return n
}

func (s *fakeRedis) exists(args []string) (string, bool) {
	return respInt(s.count(args[1:], false)), true
}

func (s *fakeRedis) del(args []string) (string, bool) {
	return respInt(s.count(args[1:], true)), true
}

func (s *fakeRedis) hset(args []string) (string, bool) {
	if len(args) < 4 || len(args)%2 != 0 {
		return "", false
	}
	h := s.hash(args[1])
	added := 0
	for i := 2; i < len(args); i += 2 {
		if _, ok := h[args[i]]; !ok {
			added++
		}
		h[args[i]] = args[i+1]
	}
	return respInt(added), true
}

func (s *fakeRedis) hget(args []string) (string, bool) {
	if len(args) != 3 {
		return "", false
	}
	v, ok := s.hashes[args[1]][args[2]]
	if !ok {
		return respNil, true
	}
	return respBulk(v), true
}

func (s *fakeRedis) hincrby(args []string) (string, bool) {
	if len(args) != 4 {
		return "", false
	}
	h := s.hash(args[1])
	cur, _ := strconv.Atoi(h[args[2]])
	inc, _ := strconv.Atoi(args[3])
	h[args[2]] = strconv.Itoa(cur + inc)
	return respInt(cur + inc), true
}

func (s *fakeRedis) hgetall(args []string) (string, bool) {
	if len(args) != 2 {
		return "", false
	}
	h := s.hashes[args[1]]
	out := fmt.Sprintf("*%d\r\n", 2*len(h))
	for k, v := range h {
		out += respBulk(k) + respBulk(v)
	}
	return out, true
}

func (s *fakeRedis) hdel(args []string) (string, bool) {
	if len(args) < 3 {
		return "", false
	}
	n := 0
	for _, f := range args[2:] {
		if _, ok := s.hashes[args[1]][f]; ok {
			delete(s.hashes[args[1]], f)
			n++
		}
	}
	return respInt(n), true
}

func (s *fakeRedis) hash(key string) map[string]string {
	if s.hashes[key] == nil {
		s.hashes[key] = map[string]string{}
	}
	return s.hashes[key]
}

// TestRedisPoolAndCommands covers tilinna/clock: radix's pool uses it for
// init timing, per-command timeouts and the periodic PING that keeps
// database.InitRedis connections healthy. It also replays the FlatCmd shapes
// used by handlers and service (JWT blacklist, recovery hashes, GETDEL).
func TestRedisPoolAndCommands(t *testing.T) {
	logIndirect(t)
	srv, host, port := startFakeRedis(t)
	cfg := setupConfig(t)
	cfg.Database.REDIS.Activate = gconfig.Activated
	cfg.Database.REDIS.Env.Host = host
	cfg.Database.REDIS.Env.Port = port
	cfg.Database.REDIS.Conn.PoolSize = 10 // default ping interval: 5s/10
	cfg.Database.REDIS.Conn.ConnTTL = 5
	cfg.Security.MustJWT = gconfig.Activated
	cfg.Security.InvalidateJWT = gconfig.Activated
	t.Cleanup(func() { _ = gdatabase.CloseRedis() })

	client, err := gdatabase.InitRedis()
	if err != nil {
		t.Fatalf("InitRedis: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// subtests run in order and share the fake's state
	t.Run("JWT blacklist", func(t *testing.T) { checkRedisBlacklist(ctx, t, client) })
	t.Run("recovery counter", func(t *testing.T) { checkRedisCounter(ctx, t, client) })
	t.Run("recovery hash", func(t *testing.T) { checkRedisHash(ctx, t, client) })
	t.Run("one-shot code", func(t *testing.T) { checkRedisOneShot(ctx, t, client) })
	t.Run("periodic PING", func(t *testing.T) { waitForRedisPing(t, srv) })
}

// checkRedisBlacklist replays handler/logout.go (SET EXAT) and
// service.IsTokenAllowed (EXISTS).
func checkRedisBlacklist(ctx context.Context, t *testing.T, client radix.Client) {
	jti := gconfig.PrefixJtiBlacklist + "jti-1"
	exp := time.Now().Add(time.Hour).Unix()
	if err := client.Do(ctx, radix.FlatCmd(nil, "SET", jti, exp, "EXAT", exp)); err != nil {
		t.Fatalf("SET EXAT: %v", err)
	}
	if gservice.IsTokenAllowed(ctx, "jti-1") {
		t.Error("blacklisted token is still allowed")
	}
	if !gservice.IsTokenAllowed(ctx, "jti-2") {
		t.Error("non-blacklisted token is not allowed")
	}
}

// checkRedisCounter replays the HINCRBY attempt counters of
// handler/passwordReset.go and service/twoFALimiter.go.
func checkRedisCounter(ctx context.Context, t *testing.T, client radix.Client) {
	var attempts int
	for want := 1; want <= 2; want++ {
		if err := client.Do(ctx, radix.FlatCmd(&attempts, "HINCRBY", "recover:1", "attempts", 1)); err != nil {
			t.Fatalf("HINCRBY: %v", err)
		}
		if attempts != want {
			t.Errorf("HINCRBY = %d, want %d", attempts, want)
		}
	}
}

// checkRedisHash replays HSET/HGETALL (example apps) and HGET/DEL
// (passwordReset.go, twoFALimiter.go) on the hash checkRedisCounter filled.
func checkRedisHash(ctx context.Context, t *testing.T, client radix.Client) {
	if err := client.Do(ctx, radix.FlatCmd(nil, "HSET", "recover:1", "email", "hash")); err != nil {
		t.Fatalf("HSET: %v", err)
	}
	var email string
	if err := client.Do(ctx, radix.FlatCmd(&email, "HGET", "recover:1", "email")); err != nil || email != "hash" {
		t.Errorf("HGET = %q, %v", email, err)
	}
	all := map[string]string{}
	if err := client.Do(ctx, radix.FlatCmd(&all, "HGETALL", "recover:1")); err != nil || all["attempts"] != "2" {
		t.Errorf("HGETALL = %v, %v", all, err)
	}
	var deleted int
	if err := client.Do(ctx, radix.FlatCmd(&deleted, "DEL", "recover:1")); err != nil || deleted != 1 {
		t.Errorf("DEL = %d, %v", deleted, err)
	}
}

// checkRedisOneShot replays the one-shot codes of handler/verification.go
// (SET, then GETDEL twice: the second must be nil).
func checkRedisOneShot(ctx context.Context, t *testing.T, client radix.Client) {
	if err := client.Do(ctx, radix.FlatCmd(nil, "SET", "verify:1", "user@example.com")); err != nil {
		t.Fatalf("SET: %v", err)
	}
	var code string
	if err := client.Do(ctx, radix.FlatCmd(&code, "GETDEL", "verify:1")); err != nil || code != "user@example.com" {
		t.Errorf("GETDEL = %q, %v", code, err)
	}
	var mb radix.Maybe
	if err := client.Do(ctx, radix.FlatCmd(&mb, "GETDEL", "verify:1")); err != nil || !mb.Null {
		t.Errorf("second GETDEL should be nil: %+v, %v", mb, err)
	}
}

// waitForRedisPing waits up to 3s for the pool's clock-driven PING, then
// checks that the fake saw no unknown command.
func waitForRedisPing(t *testing.T, srv *fakeRedis) {
	deadline := time.Now().Add(3 * time.Second)
	for {
		srv.mu.Lock()
		pings, unknown := srv.pings, srv.unknown
		srv.mu.Unlock()
		if pings > 0 {
			if len(unknown) > 0 {
				t.Errorf("unexpected commands: %q", unknown)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("radix pool never sent its periodic PING")
		}
		time.Sleep(50 * time.Millisecond)
	}
}
