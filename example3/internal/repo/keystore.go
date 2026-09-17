package repo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/pilinux/gorest/example3/internal/database/model"
)

// KeyStore is the contract for persisting the single wrapped master key.
type KeyStore interface {
	// EnsureMasterKey returns the existing master-key document or, if none
	// exists yet, inserts candidate and returns it. It never overwrites an
	// existing key, and (given the unique index created by EnsureIndexes)
	// concurrent callers cannot create a second master key.
	EnsureMasterKey(ctx context.Context, candidate *model.SecretKey) (*model.SecretKey, error)

	// UpdateWrappedKey replaces the wrapped bytes of the master-key document
	// identified by id. It is used during env-secret rotation to re-wrap the
	// (unchanged) master key with the new KEK.
	UpdateWrappedKey(ctx context.Context, id bson.ObjectID, wrappedKey []byte) error
}

// KeyStoreRepo is the MongoDB-backed KeyStore.
type KeyStoreRepo struct {
	db *mongo.Client
}

// NewKeyStoreRepo returns a new KeyStoreRepo.
func NewKeyStoreRepo(conn *mongo.Client) *KeyStoreRepo {
	return &KeyStoreRepo{db: conn}
}

// Compile-time check:
var _ KeyStore = (*KeyStoreRepo)(nil)

// coll returns the MongoDB collection that stores the wrapped master key.
func (r *KeyStoreRepo) coll() *mongo.Collection {
	return r.db.Database(DBName).Collection(CollSecretKeys)
}

// EnsureIndexes creates the unique index on keyName that the singleton
// master-key document depends on.
//
// The upsert in EnsureMasterKey is not enough on its own: MongoDB matches and
// inserts as two steps, so two instances booting against an empty collection
// can both miss the filter and both insert, leaving two master keys behind.
func (r *KeyStoreRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "keyName", Value: 1}},
		Options: options.Index().SetName("keyName_unique").SetUnique(true),
	})
	return err
}

// EnsureMasterKey implements KeyStore on top of an upsert that only ever
// inserts: $setOnInsert writes the candidate when the document does not
// exist yet, and returns the existing document when it does.
func (r *KeyStoreRepo) EnsureMasterKey(ctx context.Context, candidate *model.SecretKey) (*model.SecretKey, error) {
	stored, err := r.upsertMasterKey(ctx, candidate)
	if err != nil && mongo.IsDuplicateKeyError(err) {
		return r.upsertMasterKey(ctx, candidate)
	}
	return stored, err
}

// upsertMasterKey is one attempt at the insert-or-fetch upsert.
func (r *KeyStoreRepo) upsertMasterKey(ctx context.Context, candidate *model.SecretKey) (*model.SecretKey, error) {
	filter := bson.D{{Key: "keyName", Value: model.MasterKeyName}}
	update := bson.D{{Key: "$setOnInsert", Value: bson.D{
		{Key: "_id", Value: candidate.ID},
		{Key: "keyName", Value: candidate.KeyName},
		{Key: "wrappedKey", Value: candidate.WrappedKey},
		{Key: "createdAt", Value: candidate.CreatedAt},
		{Key: "updatedAt", Value: candidate.UpdatedAt},
	}}}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var stored model.SecretKey
	if err := r.coll().FindOneAndUpdate(ctx, filter, update, opts).Decode(&stored); err != nil {
		return nil, err
	}
	return &stored, nil
}

// UpdateWrappedKey implements KeyStore by replacing the wrapped bytes of
// the master-key document identified by id. It is used during  env-secret
// rotation to re-wrap the (unchanged) master key with the new KEK.
func (r *KeyStoreRepo) UpdateWrappedKey(ctx context.Context, id bson.ObjectID, wrappedKey []byte) error {
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "wrappedKey", Value: wrappedKey},
		{Key: "updatedAt", Value: time.Now().Unix()},
	}}}

	res, err := r.coll().UpdateOne(ctx, bson.D{{Key: "_id", Value: id}}, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
