package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// EnsureIndexes creates every index example3 relies on.
//
// It must run at boot, before the master key is loaded: the unique index on
// keyName is what makes the upsert in EnsureMasterKey safe.
//
// Creating an index is idempotent, so this is safe to call on every start: an
// index that already exists with the same specification is left alone.
func EnsureIndexes(ctx context.Context, conn *mongo.Client) error {
	if err := NewKeyStoreRepo(conn).EnsureIndexes(ctx); err != nil {
		return err
	}
	return NewFileRecordRepo(conn).EnsureIndexes(ctx)
}
