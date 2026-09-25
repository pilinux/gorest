package handler_test

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/pilinux/gorest/example3/internal/database/model"
	"github.com/pilinux/gorest/example3/internal/service"
)

// testSecret is a 40-character env secret used across the handler tests.
const testSecret = "0123456789abcdef0123456789abcdef01234567"

// init puts gin in test mode so the handlers stay quiet while testing.
func init() {
	gin.SetMode(gin.TestMode)
}

// fakeKeyStore is an in-memory repo.KeyStore: it always bootstraps a fresh key.
type fakeKeyStore struct {
	stored *model.SecretKey
}

// EnsureMasterKey keeps the first key it is given and returns it every time.
func (f *fakeKeyStore) EnsureMasterKey(_ context.Context, candidate *model.SecretKey) (*model.SecretKey, error) {
	if f.stored == nil {
		f.stored = candidate
	}
	return f.stored, nil
}

// UpdateWrappedKey rewraps the stored key, as a key rotation would.
func (f *fakeKeyStore) UpdateWrappedKey(_ context.Context, _ bson.ObjectID, wrappedKey []byte) error {
	if f.stored != nil {
		f.stored.WrappedKey = wrappedKey
	}
	return nil
}

// fakeFileStore is an in-memory repo.FileRecordStore.
type fakeFileStore struct {
	records map[string]*model.FileRecord
}

// newFakeFileStore returns an empty file store ready to use.
func newFakeFileStore() *fakeFileStore {
	return &fakeFileStore{records: make(map[string]*model.FileRecord)}
}

// Create stores the record under its file ID, giving it an ID if it has none.
func (f *fakeFileStore) Create(_ context.Context, rec *model.FileRecord) error {
	if rec.ID.IsZero() {
		rec.ID = bson.NewObjectID()
	}
	// keep only the sealed name and size
	stored := *rec
	stored.Name, stored.Size = "", 0
	f.records[rec.FileID] = &stored
	return nil
}

// GetByFileID looks up one record, or reports that no document matched.
func (f *fakeFileStore) GetByFileID(_ context.Context, fileID string) (*model.FileRecord, error) {
	rec, ok := f.records[fileID]
	if !ok {
		return nil, mongo.ErrNoDocuments
	}
	return rec, nil
}

// DeleteByFileID drops one record, or reports that no document matched.
func (f *fakeFileStore) DeleteByFileID(_ context.Context, fileID string) error {
	if _, ok := f.records[fileID]; !ok {
		return mongo.ErrNoDocuments
	}
	delete(f.records, fileID)
	return nil
}

// newLoadedKeyManager returns a KeyManager whose master key is already
// bootstrapped in memory, so no MongoDB is needed.
func newLoadedKeyManager(t *testing.T) *service.KeyManager {
	t.Helper()
	km := service.NewKeyManager()
	if err := km.SetSecrets(testSecret, ""); err != nil {
		t.Fatalf("SetSecrets error: %v", err)
	}
	if err := km.EnsureAndLoad(context.Background(), &fakeKeyStore{}); err != nil {
		t.Fatalf("EnsureAndLoad error: %v", err)
	}
	return km
}
