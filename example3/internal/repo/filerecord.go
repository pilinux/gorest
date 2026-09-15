package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/pilinux/gorest/example3/internal/database/model"
)

// FileRecordStore is the contract for the encrypted-file metadata.
type FileRecordStore interface {
	Create(ctx context.Context, rec *model.FileRecord) error
	GetByFileID(ctx context.Context, fileID string) (*model.FileRecord, error)
	DeleteByFileID(ctx context.Context, fileID string) error
}

// FileRecordRepo is the MongoDB-backed FileRecordStore.
type FileRecordRepo struct {
	db *mongo.Client
}

// NewFileRecordRepo returns a new FileRecordRepo.
func NewFileRecordRepo(conn *mongo.Client) *FileRecordRepo {
	return &FileRecordRepo{db: conn}
}

// Compile-time check:
var _ FileRecordStore = (*FileRecordRepo)(nil)

// coll returns the MongoDB collection that stores encrypted-file metadata.
func (r *FileRecordRepo) coll() *mongo.Collection {
	return r.db.Database(DBName).Collection(CollFileRecords)
}

// EnsureIndexes creates the unique index on fileId, the field every lookup and
// delete filters on.
func (r *FileRecordRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "fileId", Value: 1}},
		Options: options.Index().SetName("fileId_unique").SetUnique(true),
	})
	return err
}

// Create inserts a new file metadata document.
func (r *FileRecordRepo) Create(ctx context.Context, rec *model.FileRecord) error {
	if rec.ID.IsZero() {
		rec.ID = bson.NewObjectID()
	}
	_, err := r.coll().InsertOne(ctx, rec)
	return err
}

// GetByFileID retrieves a file metadata document by its public file id.
// It returns mongo.ErrNoDocuments when nothing matches.
func (r *FileRecordRepo) GetByFileID(ctx context.Context, fileID string) (*model.FileRecord, error) {
	var rec model.FileRecord
	err := r.coll().FindOne(ctx, bson.D{{Key: "fileId", Value: fileID}}).Decode(&rec)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// DeleteByFileID removes a file metadata document by its public file id.
// It returns mongo.ErrNoDocuments when nothing was deleted.
func (r *FileRecordRepo) DeleteByFileID(ctx context.Context, fileID string) error {
	res, err := r.coll().DeleteOne(ctx, bson.D{{Key: "fileId", Value: fileID}})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
