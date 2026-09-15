package model

import "go.mongodb.org/mongo-driver/v2/bson"

// FileRecord is the MongoDB document describing one encrypted file that lives
// on disk. The encrypted bytes themselves are stored in a file named after
// FileID; this document only keeps the metadata needed to serve the file back.
//
// Sha256 is the digest of the original (plaintext) content, recorded so a
// decryption can be verified end to end.
//
// Unpadded says how the payload was framed. Padded and unpadded files look
// alike on disk (both are 0x81 streams), so this field is the only thing that
// tells a download which opener to use. It is omitted for padded files, so the
// zero value is the padded format.
type FileRecord struct {
	ID        bson.ObjectID `json:"id" bson:"_id"`
	FileID    string        `json:"fileId" bson:"fileId"`                         // unpredictable hex id, also the on-disk filename stem
	Name      string        `json:"name" bson:"name"`                             // sanitized original filename
	Size      int64         `json:"size" bson:"size"`                             // original plaintext size in bytes
	Sha256    string        `json:"sha256" bson:"sha256"`                         // digest of the plaintext content
	Unpadded  bool          `json:"unpadded,omitempty" bson:"unpadded,omitempty"` // true if the payload was sealed without padding
	CreatedAt int64         `json:"createdAt" bson:"createdAt"`
}
