// Package model holds the MongoDB documents and the request/response DTOs used
// by example3.
package model

import "go.mongodb.org/mongo-driver/v2/bson"

// MasterKeyName is the fixed discriminator of the singleton master-key
// document. Exactly one SecretKey document with this name exists in the
// collection.
const MasterKeyName = "master"

// SecretKey is the MongoDB document that stores the master key in wrapped
// (KEK-encrypted) form. Only the wrapped bytes are ever persisted; the
// plaintext master key lives only in memory at runtime.
//
// This document must never be serialized into an API response.
type SecretKey struct {
	ID         bson.ObjectID `json:"-" bson:"_id"`
	KeyName    string        `json:"-" bson:"keyName"`    // singleton discriminator, always MasterKeyName
	WrappedKey []byte        `json:"-" bson:"wrappedKey"` // master key encrypted by the KEK
	CreatedAt  int64         `json:"-" bson:"createdAt"`
	UpdatedAt  int64         `json:"-" bson:"updatedAt"`
}
