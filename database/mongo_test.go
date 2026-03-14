package database_test

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	opts "go.mongodb.org/mongo-driver/v2/mongo/options"

	gdb "github.com/pilinux/gorest/database"
)

// TestInitMongo_Error exercises InitMongo against an invalid MongoDB URI.
// This covers the URI parsing, context creation, client options setup,
// and the error path.
func TestInitMongo_Error(t *testing.T) {
	cfg := mustGetConfig(t)
	cfg.Database.MongoDB.Env.URI = "mongodb://127.0.0.1:27099/?serverSelectionTimeoutMS=500"
	cfg.Database.MongoDB.Env.AppName = "gorest-test"
	cfg.Database.MongoDB.Env.PoolSize = 1
	cfg.Database.MongoDB.Env.ConnTTL = 2
	cfg.Database.MongoDB.Env.PoolMon = ""

	client, err := gdb.InitMongo()
	if err == nil {
		t.Fatal("expected error from InitMongo with no server, got nil")
	}
	if client != nil {
		t.Log("InitMongo returned a non-nil client with error")
	}
}

// TestInitMongo_PoolMonitor exercises InitMongo with pool monitoring enabled.
// The connection will still fail (no server), but the pool monitor setup
// code path is covered.
func TestInitMongo_PoolMonitor(t *testing.T) {
	cfg := mustGetConfig(t)
	cfg.Database.MongoDB.Env.URI = "mongodb://127.0.0.1:27099/?serverSelectionTimeoutMS=500"
	cfg.Database.MongoDB.Env.AppName = "gorest-test"
	cfg.Database.MongoDB.Env.PoolSize = 1
	cfg.Database.MongoDB.Env.ConnTTL = 2
	cfg.Database.MongoDB.Env.PoolMon = "yes"

	client, err := gdb.InitMongo()
	if err == nil {
		t.Fatal("expected error from InitMongo with pool monitor, got nil")
	}
	if client != nil {
		t.Log("InitMongo returned a non-nil client with error")
	}

	// reset pool monitor
	cfg.Database.MongoDB.Env.PoolMon = ""
}

// TestInitMongo_InvalidURI tests InitMongo with a completely invalid URI
// that causes mongo.Connect to fail before even attempting a ping.
func TestInitMongo_InvalidURI(t *testing.T) {
	cfg := mustGetConfig(t)
	cfg.Database.MongoDB.Env.URI = "://invalid"
	cfg.Database.MongoDB.Env.AppName = "gorest-test"
	cfg.Database.MongoDB.Env.PoolSize = 1
	cfg.Database.MongoDB.Env.ConnTTL = 2
	cfg.Database.MongoDB.Env.PoolMon = ""

	client, err := gdb.InitMongo()
	if err == nil {
		t.Fatal("expected error from InitMongo with invalid URI, got nil")
	}
	if client != nil {
		t.Log("InitMongo returned a non-nil client with error")
	}
}

// TestGetMongo_Nil verifies GetMongo returns nil when no MongoDB
// connection has been established.
func TestGetMongo_Nil(t *testing.T) {
	_ = gdb.CloseMongo()

	client := gdb.GetMongo()
	if client != nil {
		t.Fatal("expected nil from GetMongo when no connection exists")
	}
}

// newTestMongoClient creates a *mongo.Client via mongo.Connect using a
// bogus URI. The client is valid but not connected to any server, which
// is sufficient for testing Disconnect behavior.
func newTestMongoClient(t *testing.T) *mongo.Client {
	t.Helper()
	clientOptions := opts.Client().
		ApplyURI("mongodb://127.0.0.1:27099/").
		SetAppName("gorest-test").
		SetMaxPoolSize(1).
		SetConnectTimeout(1 * time.Second)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		t.Fatalf("mongo.Connect failed: %v", err)
	}
	return client
}

// TestCloseMongo_Success injects a *mongo.Client that has not been
// disconnected yet. Calling CloseMongo should succeed.
func TestCloseMongo_Success(t *testing.T) {
	client := newTestMongoClient(t)
	gdb.SetMongoClient(client)
	defer gdb.ResetMongoClient()

	if err := gdb.CloseMongo(); err != nil {
		t.Fatalf("expected nil error from CloseMongo, got: %v", err)
	}

	// after successful close, GetMongo should return nil.
	if gdb.GetMongo() != nil {
		t.Fatal("expected GetMongo to return nil after CloseMongo")
	}
}

// TestCloseMongo_DisconnectError injects a *mongo.Client that has
// already been disconnected. The second Disconnect call inside
// CloseMongo will return an error.
func TestCloseMongo_DisconnectError(t *testing.T) {
	client := newTestMongoClient(t)

	// Disconnect once so the client enters the disconnected state.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Disconnect(ctx); err != nil {
		t.Fatalf("first Disconnect failed unexpectedly: %v", err)
	}

	gdb.SetMongoClient(client)
	defer gdb.ResetMongoClient()

	err := gdb.CloseMongo()
	if err == nil {
		t.Fatal("expected error from CloseMongo on already-disconnected client, got nil")
	}
	t.Logf("CloseMongo returned expected error: %v", err)
}
