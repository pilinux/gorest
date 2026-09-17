// Package repo provides the MongoDB persistence layer for example3: the wrapped
// master key and the encrypted-file metadata. Every concrete repository is
// hidden behind an interface.
package repo

// DBName is the MongoDB database used by example3.
const DBName = "example3"

// Collection names.
const (
	CollSecretKeys  = "secretKeys"
	CollFileRecords = "fileRecords"
)
