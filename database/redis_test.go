package database_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/mediocregopher/radix/v4"
	gdb "github.com/pilinux/gorest/database"
)

// TestInitRedis_Error exercises InitRedis against a non-existent Redis
// server. This covers the DSN building, context creation, and the
// error-handling return path.
func TestInitRedis_Error(t *testing.T) {
	cfg := mustGetConfig(t)
	cfg.Database.REDIS.Env.Host = "127.0.0.1"
	cfg.Database.REDIS.Env.Port = "16379" // unlikely to be running
	cfg.Database.REDIS.Conn.PoolSize = 1
	cfg.Database.REDIS.Conn.ConnTTL = 2

	client, err := gdb.InitRedis()
	if err == nil {
		t.Fatal("expected error from InitRedis with no server, got nil")
	}
	if client != nil {
		// if a partial client was returned, try to close it
		t.Log("InitRedis returned a non-nil client with error; closing")
	}
}

// TestGetRedis_Nil verifies GetRedis returns nil when no Redis
// connection has been established.
func TestGetRedis_Nil(t *testing.T) {
	// ensure redis is closed from any previous test
	_ = gdb.CloseRedis()

	client := gdb.GetRedis()
	if client != nil {
		t.Fatal("expected nil from GetRedis when no connection exists")
	}
}

// mockRedisClient implements radix.Client for testing CloseRedis.
type mockRedisClient struct {
	closeErr error
}

func (m *mockRedisClient) Addr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 6379}
}

func (m *mockRedisClient) Do(_ context.Context, _ radix.Action) error {
	return nil
}

func (m *mockRedisClient) Close() error {
	return m.closeErr
}

// TestCloseRedis_Success injects a mock radix.Client whose Close()
// succeeds, and verifying that redisClient is set to nil afterward.
func TestCloseRedis_Success(t *testing.T) {
	gdb.SetRedisClient(&mockRedisClient{closeErr: nil})
	defer gdb.ResetRedisClient()

	if err := gdb.CloseRedis(); err != nil {
		t.Fatalf("expected nil error from CloseRedis, got: %v", err)
	}

	// After successful close, GetRedis should return nil.
	if gdb.GetRedis() != nil {
		t.Fatal("expected GetRedis to return nil after CloseRedis")
	}
}

// TestCloseRedis_CloseError injects a mock radix.Client whose Close()
// returns an error, and verifying that CloseRedis propagates the error.
func TestCloseRedis_CloseError(t *testing.T) {
	errClose := errors.New("mock redis close failure")
	gdb.SetRedisClient(&mockRedisClient{closeErr: errClose})
	defer gdb.ResetRedisClient()

	err := gdb.CloseRedis()
	if err == nil {
		t.Fatal("expected error from CloseRedis, got nil")
	}
	if !errors.Is(err, errClose) {
		t.Fatalf("expected error %q, got %q", errClose, err)
	}
}
