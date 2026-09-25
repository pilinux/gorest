package model

import "go.mongodb.org/mongo-driver/v2/bson"

// FileRecord is the MongoDB document describing one encrypted file that lives
// on disk. The encrypted bytes themselves are stored in a file named after
// FileID; this document only keeps the metadata needed to serve the file back.
//
// Name and Size are never stored as they are. Only their sealed copies go to
// the database.
//
// Unpadded says how the payload was framed. Padded and unpadded files look
// alike on disk (both are 0x81 streams), so this field is the only thing that
// tells a download which opener to use. It is omitted for padded files, so the
// zero value is the padded format.
type FileRecord struct {
	ID         bson.ObjectID `json:"id" bson:"_id"`
	FileID     string        `json:"fileId" bson:"fileId"`                         // unpredictable hex id, also the on-disk filename stem
	Name       string        `json:"name" bson:"-"`                                // sanitized original filename
	Size       int64         `json:"size" bson:"-"`                                // original plaintext size in bytes
	SealedName string        `json:"-" bson:"name"`                                // Name, sealed to FileID
	SealedSize string        `json:"-" bson:"size"`                                // Size, sealed to FileID
	Unpadded   bool          `json:"unpadded,omitempty" bson:"unpadded,omitempty"` // true if the payload was sealed without padding
	CreatedAt  int64         `json:"createdAt" bson:"createdAt"`
}
