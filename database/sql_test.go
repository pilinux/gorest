package database_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	gdb "github.com/pilinux/gorest/database"

	"gorm.io/gorm"
)

// TestInitDB_Sqlite3 tests InitDB with the sqlite3 driver using an
// in-memory database, then verifies GetDB returns a valid connection
// and CloseSQL cleanly shuts it down.
func TestInitDB_Sqlite3(t *testing.T) {
	cfg := mustGetConfig(t)
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	cfg.Database.RDBMS.Log.LogLevel = 1

	db := gdb.InitDB()
	if db == nil {
		t.Fatal("InitDB returned nil")
	}
	if db.Error != nil {
		t.Fatalf("InitDB error: %v", db.Error)
	}

	// GetDB should return the same instance
	got := gdb.GetDB()
	if got == nil {
		t.Fatal("GetDB returned nil after successful InitDB")
	}

	// CloseSQL should close without error
	if err := gdb.CloseSQL(); err != nil {
		t.Fatalf("CloseSQL error: %v", err)
	}

	// after close, GetDB may still hold a reference but the underlying
	// SQL connection is closed; calling CloseSQL again should be safe
	if err := gdb.CloseSQL(); err != nil {
		t.Fatalf("second CloseSQL error: %v", err)
	}
}

// TestInitDB_DefaultDriver tests InitDB with an unsupported driver name.
// The function must have been initialised with a valid db first so that
// the internal dbClient is non-nil (otherwise setting db.Error panics).
func TestInitDB_DefaultDriver(t *testing.T) {
	// first, init a valid sqlite3 connection so dbClient is non-nil
	cfg := mustGetConfig(t)
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	cfg.Database.RDBMS.Log.LogLevel = 1

	db := gdb.InitDB()
	if db == nil || db.Error != nil {
		t.Fatalf("prerequisite InitDB failed: %v", db.Error)
	}

	// now switch to an unsupported driver
	cfg.Database.RDBMS.Env.Driver = "unsupported"

	db = gdb.InitDB()
	if db == nil {
		t.Fatal("InitDB returned nil for unsupported driver")
	}
	if db.Error == nil {
		t.Fatal("expected error for unsupported driver, got nil")
	}

	// clean up
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	if err := gdb.CloseSQL(); err != nil {
		t.Logf("CloseSQL during cleanup: %v", err)
	}
}

// TestInitDB_MysqlDSN exercises the MySQL branch of InitDB.
// Without a running MySQL server, sql.Open succeeds but gorm.Open
// fails on the connection attempt, covering the DSN building and
// error-handling code.
func TestInitDB_MysqlDSN(t *testing.T) {
	cfg := mustGetConfig(t)

	// first, ensure dbClient is non-nil via sqlite3
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	db := gdb.InitDB()
	if db == nil || db.Error != nil {
		t.Fatalf("prerequisite InitDB failed: %v", db.Error)
	}

	tests := []struct {
		name    string
		host    string
		port    string
		sslmode string
	}{
		{
			name:    "mysql default",
			host:    "127.0.0.1",
			port:    "13306",
			sslmode: "",
		},
		{
			name:    "mysql with port and sslmode disable",
			host:    "127.0.0.1",
			port:    "13306",
			sslmode: "disable",
		},
		{
			name:    "mysql sslmode require",
			host:    "127.0.0.1",
			port:    "13306",
			sslmode: "require",
		},
		{
			name:    "mysql no port",
			host:    "127.0.0.1",
			port:    "",
			sslmode: "disable",
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			cfg.Database.RDBMS.Env.Driver = "mysql"
			cfg.Database.RDBMS.Env.Host = tc.host
			cfg.Database.RDBMS.Env.Port = tc.port
			cfg.Database.RDBMS.Access.User = "test_user"
			cfg.Database.RDBMS.Access.Pass = "test_pass"
			cfg.Database.RDBMS.Access.DbName = "test_db"
			cfg.Database.RDBMS.Ssl.Sslmode = tc.sslmode
			cfg.Database.RDBMS.Conn.MaxIdleConns = 2
			cfg.Database.RDBMS.Conn.MaxOpenConns = 5
			cfg.Database.RDBMS.Log.LogLevel = 1

			db := gdb.InitDB()
			if db == nil {
				t.Fatal("InitDB returned nil")
			}
			// gorm.Open will fail connecting, which is expected
			if db.Error == nil {
				// if it somehow connected, clean up
				_ = gdb.CloseSQL()
			}
		})
	}

	// restore sqlite3 for cleanup
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	_ = gdb.CloseSQL()
}

// TestInitDB_PostgresDSN exercises the Postgres branch of InitDB.
func TestInitDB_PostgresDSN(t *testing.T) {
	cfg := mustGetConfig(t)

	// ensure dbClient is non-nil
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	db := gdb.InitDB()
	if db == nil || db.Error != nil {
		t.Fatalf("prerequisite InitDB failed: %v", db.Error)
	}

	tests := []struct {
		name       string
		host       string
		port       string
		sslmode    string
		rootCA     string
		serverCert string
		clientCert string
		clientKey  string
	}{
		{
			name:    "postgres default",
			host:    "127.0.0.1",
			port:    "15432",
			sslmode: "",
		},
		{
			name:    "postgres with port and ssl disable",
			host:    "127.0.0.1",
			port:    "15432",
			sslmode: "disable",
		},
		{
			name:    "postgres no port",
			host:    "127.0.0.1",
			port:    "",
			sslmode: "disable",
		},
		{
			name:       "postgres ssl with rootCA",
			host:       "127.0.0.1",
			port:       "15432",
			sslmode:    "verify-full",
			rootCA:     "/nonexistent/ca.pem",
			serverCert: "",
			clientCert: "/nonexistent/client.pem",
			clientKey:  "/nonexistent/client-key.pem",
		},
		{
			name:       "postgres ssl with serverCert",
			host:       "127.0.0.1",
			port:       "15432",
			sslmode:    "require",
			rootCA:     "",
			serverCert: "/nonexistent/server.pem",
		},
		{
			name:       "postgres ssl with clientCert only",
			host:       "127.0.0.1",
			port:       "15432",
			sslmode:    "require",
			clientCert: "/nonexistent/client.pem",
		},
		{
			name:      "postgres ssl with clientKey only",
			host:      "127.0.0.1",
			port:      "15432",
			sslmode:   "require",
			clientKey: "/nonexistent/client-key.pem",
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			cfg.Database.RDBMS.Env.Driver = "postgres"
			cfg.Database.RDBMS.Env.Host = tc.host
			cfg.Database.RDBMS.Env.Port = tc.port
			cfg.Database.RDBMS.Env.TimeZone = "UTC"
			cfg.Database.RDBMS.Access.User = "test_user"
			cfg.Database.RDBMS.Access.Pass = "test_pass"
			cfg.Database.RDBMS.Access.DbName = "test_db"
			cfg.Database.RDBMS.Ssl.Sslmode = tc.sslmode
			cfg.Database.RDBMS.Ssl.RootCA = tc.rootCA
			cfg.Database.RDBMS.Ssl.ServerCert = tc.serverCert
			cfg.Database.RDBMS.Ssl.ClientCert = tc.clientCert
			cfg.Database.RDBMS.Ssl.ClientKey = tc.clientKey
			cfg.Database.RDBMS.Conn.MaxIdleConns = 2
			cfg.Database.RDBMS.Conn.MaxOpenConns = 5
			cfg.Database.RDBMS.Log.LogLevel = 1

			db := gdb.InitDB()
			if db == nil {
				t.Fatal("InitDB returned nil")
			}
			if db.Error == nil {
				_ = gdb.CloseSQL()
			}
		})
	}

	// restore and cleanup
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	cfg.Database.RDBMS.Ssl.Sslmode = ""
	cfg.Database.RDBMS.Ssl.RootCA = ""
	cfg.Database.RDBMS.Ssl.ServerCert = ""
	cfg.Database.RDBMS.Ssl.ClientCert = ""
	cfg.Database.RDBMS.Ssl.ClientKey = ""
	_ = gdb.CloseSQL()
}

// TestInitDB_MysqlVerifyCA exercises the MySQL verify-ca sslmode branch.
// InitTLSMySQL is expected to fail because no valid certs are configured,
// which covers the TLS-error return path inside InitDB.
func TestInitDB_MysqlVerifyCA(t *testing.T) {
	cfg := mustGetConfig(t)

	// ensure dbClient is non-nil
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	db := gdb.InitDB()
	if db == nil || db.Error != nil {
		t.Fatalf("prerequisite InitDB failed: %v", db.Error)
	}

	tests := []struct {
		name    string
		sslmode string
	}{
		{name: "verify-ca", sslmode: "verify-ca"},
		{name: "verify-full", sslmode: "verify-full"},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			cfg.Database.RDBMS.Env.Driver = "mysql"
			cfg.Database.RDBMS.Env.Host = "127.0.0.1"
			cfg.Database.RDBMS.Env.Port = "13306"
			cfg.Database.RDBMS.Access.User = "test_user"
			cfg.Database.RDBMS.Access.Pass = "test_pass"
			cfg.Database.RDBMS.Access.DbName = "test_db"
			cfg.Database.RDBMS.Ssl.Sslmode = tc.sslmode
			// no valid certs, so InitTLSMySQL will fail
			cfg.Database.RDBMS.Ssl.RootCA = ""
			cfg.Database.RDBMS.Ssl.ServerCert = ""
			cfg.Database.RDBMS.Ssl.ClientCert = ""
			cfg.Database.RDBMS.Ssl.ClientKey = ""
			cfg.Database.RDBMS.Conn.MaxIdleConns = 2
			cfg.Database.RDBMS.Conn.MaxOpenConns = 5
			cfg.Database.RDBMS.Log.LogLevel = 1

			db := gdb.InitDB()
			if db == nil {
				t.Fatal("InitDB returned nil")
			}
			if db.Error == nil {
				_ = gdb.CloseSQL()
				t.Fatal("expected error for mysql verify-ca without certs, got nil")
			}
		})
	}

	// cleanup
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	cfg.Database.RDBMS.Ssl.Sslmode = ""
	_ = gdb.CloseSQL()
}

// TestInitDB_Sqlite3_LazyPath verifies that sqlite3 with a non-existent
// directory path still succeeds on gorm.Open (sqlite3 is lazy). This
// exercises the success path of the sqlite3 branch.
func TestInitDB_Sqlite3_LazyPath(t *testing.T) {
	cfg := mustGetConfig(t)

	// ensure dbClient is non-nil first
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	db := gdb.InitDB()
	if db == nil || db.Error != nil {
		t.Fatalf("prerequisite InitDB failed: %v", db.Error)
	}

	// sqlite3 is lazy — gorm.Open succeeds even with a bad path
	cfg.Database.RDBMS.Access.DbName = "/nonexistent/dir/db.sqlite3"
	cfg.Database.RDBMS.Log.LogLevel = 1

	db = gdb.InitDB()
	if db == nil {
		t.Fatal("InitDB returned nil")
	}
	// sqlite3 is permissive; error occurs only on first query
	// so we just verify InitDB returned without crashing

	// cleanup
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	_ = gdb.CloseSQL()
}

// TestInitDB_MysqlSqlOpenError tests that InitDB returns an error
// when sql.Open fails for the MySQL driver. A database name containing
// an invalid percent-escape causes mysql.ParseDSN to fail.
func TestInitDB_MysqlSqlOpenError(t *testing.T) {
	cfg := mustGetConfig(t)

	// ensure dbClient is non-nil
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	db := gdb.InitDB()
	if db == nil || db.Error != nil {
		t.Fatalf("prerequisite InitDB failed: %v", db.Error)
	}

	// switch to mysql with a database name that makes ParseDSN fail
	cfg.Database.RDBMS.Env.Driver = "mysql"
	cfg.Database.RDBMS.Env.Host = "127.0.0.1"
	cfg.Database.RDBMS.Env.Port = "13306"
	cfg.Database.RDBMS.Access.User = "test_user"
	cfg.Database.RDBMS.Access.Pass = "test_pass"
	cfg.Database.RDBMS.Access.DbName = "db%xx_name" // invalid percent-escape
	cfg.Database.RDBMS.Ssl.Sslmode = "disable"
	cfg.Database.RDBMS.Conn.MaxIdleConns = 2
	cfg.Database.RDBMS.Conn.MaxOpenConns = 5
	cfg.Database.RDBMS.Log.LogLevel = 1

	db = gdb.InitDB()
	if db == nil {
		t.Fatal("InitDB returned nil")
	}
	if db.Error == nil {
		_ = gdb.CloseSQL()
		t.Fatal("expected error for invalid MySQL DSN, got nil")
	}
	if !strings.Contains(db.Error.Error(), "failed to open SQL connection") {
		t.Fatalf("unexpected error: %v", db.Error)
	}

	// cleanup
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	_ = gdb.CloseSQL()
}

// TestInitDB_PostgresSqlOpenError tests that InitDB returns an error
// when sqlOpen fails for the Postgres driver by injecting a failing
// sqlOpen function.
func TestInitDB_PostgresSqlOpenError(t *testing.T) {
	cfg := mustGetConfig(t)

	// ensure dbClient is non-nil
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	db := gdb.InitDB()
	if db == nil || db.Error != nil {
		t.Fatalf("prerequisite InitDB failed: %v", db.Error)
	}

	// inject a failing sqlOpen
	gdb.SetSQLOpen(func(_ string, _ string) (*sql.DB, error) {
		return nil, errors.New("injected sql.Open error")
	})
	defer gdb.ResetSQLOpen()

	cfg.Database.RDBMS.Env.Driver = "postgres"
	cfg.Database.RDBMS.Env.Host = "127.0.0.1"
	cfg.Database.RDBMS.Env.Port = "15432"
	cfg.Database.RDBMS.Env.TimeZone = "UTC"
	cfg.Database.RDBMS.Access.User = "test_user"
	cfg.Database.RDBMS.Access.Pass = "test_pass"
	cfg.Database.RDBMS.Access.DbName = "test_db"
	cfg.Database.RDBMS.Ssl.Sslmode = "disable"
	cfg.Database.RDBMS.Conn.MaxIdleConns = 2
	cfg.Database.RDBMS.Conn.MaxOpenConns = 5
	cfg.Database.RDBMS.Log.LogLevel = 1

	db = gdb.InitDB()
	if db == nil {
		t.Fatal("InitDB returned nil")
	}
	if db.Error == nil {
		_ = gdb.CloseSQL()
		t.Fatal("expected error for injected sql.Open failure, got nil")
	}
	if !strings.Contains(db.Error.Error(), "failed to open SQL connection") {
		t.Fatalf("unexpected error: %v", db.Error)
	}

	// cleanup
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = ":memory:"
	cfg.Database.RDBMS.Ssl.Sslmode = ""
	_ = gdb.CloseSQL()
}

// TestCloseSQL_DBMethodError tests that CloseSQL returns an error
// when dbClient.DB() fails. This happens when dbClient has a nil
// ConnPool (a bare *gorm.DB with an empty Config).
func TestCloseSQL_DBMethodError(t *testing.T) {
	// set dbClient to a gorm.DB with a non-nil Config but nil ConnPool;
	// gorm.DB embeds *Config, so Config must be non-nil to avoid panic
	gdb.SetDBClient(&gorm.DB{Config: &gorm.Config{}})
	defer gdb.ResetDBClient()

	err := gdb.CloseSQL()
	if err == nil {
		t.Fatal("expected error from CloseSQL when DB() fails, got nil")
	}
}

// failCloseDriver implements driver.Connector and produces connections
// whose Close() always returns an error.
type failCloseDriver struct{}

func (f failCloseDriver) Connect(_ context.Context) (driver.Conn, error) {
	return failCloseConn{}, nil
}

func (f failCloseDriver) Driver() driver.Driver {
	return nil
}

// failCloseConn is a minimal driver.Conn whose Close() always fails.
type failCloseConn struct{}

func (f failCloseConn) Prepare(_ string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func (f failCloseConn) Close() error {
	return errors.New("forced close error")
}

func (f failCloseConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not implemented")
}

// TestCloseSQL_CloseError tests that CloseSQL returns an error when
// the underlying sql.DB.Close() fails.
func TestCloseSQL_CloseError(t *testing.T) {
	// create a *sql.DB with a connector that produces fail-close connections
	sqlDB := sql.OpenDB(failCloseDriver{})

	// ping to establish an idle connection in the pool
	if err := sqlDB.Ping(); err != nil {
		// Ping calls Connect, which returns failCloseConn; the driver
		// doesn't implement driver.Pinger so the default path calls
		// Prepare, which fails. But the connection is still created.
		// We need to check if a connection was created despite the error.
		_ = err
	}

	// wrap in a gorm.DB and set as dbClient
	gormDB := &gorm.DB{Config: &gorm.Config{}}
	gormDB.ConnPool = sqlDB
	gdb.SetDBClient(gormDB)
	defer gdb.ResetDBClient()

	err := gdb.CloseSQL()
	if err == nil {
		t.Fatal("expected error from CloseSQL when db.Close() fails, got nil")
	}
}
