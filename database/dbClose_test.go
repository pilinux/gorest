package database_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/mediocregopher/radix/v4"
	gdb "github.com/pilinux/gorest/database"
	"gorm.io/gorm"
)

// okCloseConn is a minimal driver.Conn whose Close() succeeds.
type okCloseConn struct{}

func (okCloseConn) Prepare(_ string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func (okCloseConn) Close() error {
	return nil
}

func (okCloseConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not implemented")
}

// okCloseDriver yields connections that close cleanly.
type okCloseDriver struct{}

func (okCloseDriver) Connect(_ context.Context) (driver.Conn, error) {
	return okCloseConn{}, nil
}

func (okCloseDriver) Driver() driver.Driver {
	return nil
}

// liveSQLClient returns a *gorm.DB backed by a real *sql.DB that closes
// successfully, so CloseSQL exercises its full close path.
func liveSQLClient() *gorm.DB {
	sqlDB := sql.OpenDB(okCloseDriver{})
	// establish an idle connection so Close() actually closes a conn
	_ = sqlDB.Ping()

	gormDB := &gorm.DB{Config: &gorm.Config{}}
	gormDB.ConnPool = sqlDB
	return gormDB
}

// TestCloseSQL_Concurrent ensures concurrent CloseSQL calls with no open
// connection are a race-free no-op (double-close protection).
func TestCloseSQL_Concurrent(t *testing.T) {
	t.Log("running:" + t.Name())
	const goroutines = 50
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- gdb.CloseSQL()
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("CloseSQL returned error: %v", err)
		}
	}
}

// TestCloseRedis_Concurrent ensures concurrent CloseRedis calls with no open
// connection are a race-free no-op (double-close protection).
func TestCloseRedis_Concurrent(t *testing.T) {
	t.Log("running:" + t.Name())
	const goroutines = 50
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- gdb.CloseRedis()
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("CloseRedis returned error: %v", err)
		}
	}
}

// TestCloseMongo_Concurrent ensures concurrent CloseMongo calls with no open
// connection are a race-free no-op (double-close protection).
func TestCloseMongo_Concurrent(t *testing.T) {
	t.Log("running:" + t.Name())
	const goroutines = 50
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- gdb.CloseMongo()
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("CloseMongo returned error: %v", err)
		}
	}
}

// TestCloseAllDB_Concurrent ensures CloseAllDB closes open connections
// exactly once under concurrent callers and is race-free. Open (mock)
// connections are injected so the closers mutate shared state and the
// mutex is actually exercised, not just the nil fast path. Run with -race.
func TestCloseAllDB_Concurrent(t *testing.T) {
	t.Log("running:" + t.Name())

	gdb.ResetCloseAllOnce()
	defer gdb.ResetCloseAllOnce()

	gdb.SetDBClient(liveSQLClient())
	defer gdb.ResetDBClient()
	gdb.SetRedisClient(&closeAllDBRedisErrClient{closeErr: nil})
	defer gdb.ResetRedisClient()
	gdb.SetMongoClient(newTestMongoClient(t))
	defer gdb.ResetMongoClient()

	const goroutines = 50
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- gdb.CloseAllDB()
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("CloseAllDB returned error: %v", err)
		}
	}

	// every connection must have been closed exactly once
	if gdb.GetDB() != nil {
		t.Error("expected dbClient to be nil after CloseAllDB")
	}
	if gdb.GetRedis() != nil {
		t.Error("expected redisClient to be nil after CloseAllDB")
	}
	if gdb.GetMongo() != nil {
		t.Error("expected mongoClient to be nil after CloseAllDB")
	}
}

// closeAllDBRedisErrClient implements radix.Client with a failing Close().
type closeAllDBRedisErrClient struct {
	closeErr error
}

func (m *closeAllDBRedisErrClient) Addr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 6379}
}

func (m *closeAllDBRedisErrClient) Do(_ context.Context, _ radix.Action) error {
	return nil
}

func (m *closeAllDBRedisErrClient) Close() error {
	return m.closeErr
}

// TestCloseAllDB_CloseSQLError exercises CloseAllDB when CloseSQL
// returns an error; the aggregated error must be propagated.
func TestCloseAllDB_CloseSQLError(t *testing.T) {
	gdb.ResetCloseAllOnce()
	defer gdb.ResetCloseAllOnce()

	// inject a gorm.DB that will fail on DB() call
	gdb.SetDBClient(&gorm.DB{Config: &gorm.Config{}})
	defer gdb.ResetDBClient()

	err := gdb.CloseAllDB()
	if err == nil {
		t.Fatal("expected error from CloseAllDB when CloseSQL fails, got nil")
	}
}

// TestCloseAllDB_CloseRedisError exercises CloseAllDB when CloseRedis
// returns an error; the aggregated error must be propagated.
func TestCloseAllDB_CloseRedisError(t *testing.T) {
	gdb.ResetCloseAllOnce()
	defer gdb.ResetCloseAllOnce()

	// SQL is nil (CloseSQL succeeds), but Redis will fail
	gdb.ResetDBClient()
	gdb.SetRedisClient(&closeAllDBRedisErrClient{closeErr: errors.New("redis close fail")})
	defer gdb.ResetRedisClient()

	err := gdb.CloseAllDB()
	if err == nil {
		t.Fatal("expected error from CloseAllDB when CloseRedis fails, got nil")
	}
}

// TestCloseAllDB_CloseMongoError exercises CloseAllDB when CloseMongo
// returns an error; the aggregated error must be propagated.
func TestCloseAllDB_CloseMongoError(t *testing.T) {
	gdb.ResetCloseAllOnce()
	defer gdb.ResetCloseAllOnce()

	// SQL nil, Redis nil, Mongo fails via double-disconnect
	gdb.ResetDBClient()
	gdb.ResetRedisClient()

	// create a disconnected Mongo client
	client := newTestMongoClient(t)
	// first disconnect succeeds
	if err := client.Disconnect(context.Background()); err != nil {
		t.Fatalf("first disconnect failed: %v", err)
	}
	// inject the already-disconnected client so CloseMongo will fail
	gdb.SetMongoClient(client)
	defer gdb.ResetMongoClient()

	closeErr := gdb.CloseAllDB()
	if closeErr == nil {
		t.Fatal("expected error from CloseAllDB when CloseMongo fails, got nil")
	}
}

// TestCloseAllDB_AllSuccess exercises CloseAllDB when all connections
// are nil (all close operations succeed).
func TestCloseAllDB_AllSuccess(t *testing.T) {
	gdb.ResetCloseAllOnce()
	defer gdb.ResetCloseAllOnce()

	gdb.ResetDBClient()
	gdb.ResetRedisClient()
	gdb.ResetMongoClient()

	err := gdb.CloseAllDB()
	if err != nil {
		t.Fatalf("expected nil from CloseAllDB, got: %v", err)
	}
}

// TestCloseAllDB_AggregatesAndRetries verifies that a failing closer does
// not prevent the others from closing, and that a later call retries the
// connections that are still open (i.e. a failure is not a permanent no-op).
func TestCloseAllDB_AggregatesAndRetries(t *testing.T) {
	gdb.ResetCloseAllOnce()
	defer gdb.ResetCloseAllOnce()

	// SQL fails (gorm.DB with no conn pool); Redis and Mongo succeed.
	gdb.SetDBClient(&gorm.DB{Config: &gorm.Config{}})
	defer gdb.ResetDBClient()
	gdb.SetRedisClient(&closeAllDBRedisErrClient{closeErr: nil})
	defer gdb.ResetRedisClient()
	gdb.SetMongoClient(newTestMongoClient(t))
	defer gdb.ResetMongoClient()

	// first call: CloseSQL fails, but Redis and Mongo must still be closed
	if err := gdb.CloseAllDB(); err == nil {
		t.Fatal("expected error from CloseAllDB when CloseSQL fails, got nil")
	}
	if gdb.GetRedis() != nil {
		t.Error("Redis must be closed even though CloseSQL failed")
	}
	if gdb.GetMongo() != nil {
		t.Error("Mongo must be closed even though CloseSQL failed")
	}

	// resolve the SQL failure; a retry must now succeed and not be a no-op
	gdb.ResetDBClient()
	if err := gdb.CloseAllDB(); err != nil {
		t.Fatalf("expected retry CloseAllDB to succeed, got: %v", err)
	}
}
