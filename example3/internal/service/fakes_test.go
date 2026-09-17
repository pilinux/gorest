package service

import (
	"context"
	"io"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/pilinux/gorest/example3/internal/database/model"
)

// stubKeys is a masterKeyProvider backed by a fixed key (or a fixed error). It
// lets the crypto services be tested without a real KeyManager or database.
type stubKeys struct {
	key []byte
	err error
}

func (s stubKeys) MasterKey() ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.key, nil
}

// fakeKeyStore is an in-memory repo.KeyStore for KeyManager tests.
type fakeKeyStore struct {
	stored      *model.SecretKey
	ensureErr   error
	updateErr   error
	updateCalls int
}

func (f *fakeKeyStore) EnsureMasterKey(_ context.Context, candidate *model.SecretKey) (*model.SecretKey, error) {
	if f.ensureErr != nil {
		return nil, f.ensureErr
	}
	// fresh "database": the candidate becomes the stored master key
	if f.stored == nil {
		f.stored = candidate
	}
	return f.stored, nil
}

func (f *fakeKeyStore) UpdateWrappedKey(_ context.Context, _ bson.ObjectID, wrappedKey []byte) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updateCalls++
	if f.stored != nil {
		f.stored.WrappedKey = wrappedKey
	}
	return nil
}

// fakeFileStore is an in-memory repo.FileRecordStore for FileCryptService tests.
type fakeFileStore struct {
	records   map[string]*model.FileRecord
	createErr error
	getErr    error
	deleteErr error
}

func newFakeFileStore() *fakeFileStore {
	return &fakeFileStore{records: make(map[string]*model.FileRecord)}
}

func (f *fakeFileStore) Create(_ context.Context, rec *model.FileRecord) error {
	if f.createErr != nil {
		return f.createErr
	}
	if rec.ID.IsZero() {
		rec.ID = bson.NewObjectID()
	}
	f.records[rec.FileID] = rec
	return nil
}

func (f *fakeFileStore) GetByFileID(_ context.Context, fileID string) (*model.FileRecord, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	rec, ok := f.records[fileID]
	if !ok {
		return nil, mongo.ErrNoDocuments
	}
	return rec, nil
}

func (f *fakeFileStore) DeleteByFileID(_ context.Context, fileID string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	if _, ok := f.records[fileID]; !ok {
		return mongo.ErrNoDocuments
	}
	delete(f.records, fileID)
	return nil
}

// errReader is an io.Reader that always fails, standing in for an upload whose
// client disappears mid-transfer.
type errReader struct {
	err error
}

func (e errReader) Read(_ []byte) (int, error) {
	return 0, e.err
}

// hookReader runs fn once, as soon as at bytes have been read through it. It
// lets a test inspect the disk at a chosen moment part-way through an upload,
// while the stored file is open and only half written.
type hookReader struct {
	r    io.Reader
	at   int64
	fn   func()
	n    int64
	done bool
}

func (h *hookReader) Read(p []byte) (int, error) {
	n, err := h.r.Read(p)
	h.n += int64(n)
	if !h.done && h.n >= h.at {
		h.done = true
		h.fn()
	}
	return n, err
}
