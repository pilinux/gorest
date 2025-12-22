package lib

import (
	"golang.org/x/crypto/argon2"
)

// GetArgon2Key derives a key from the password and salt using Argon2id.
// It uses fixed parameters: time=2, memory=64MB, threads=2.
func GetArgon2Key(password []byte, salt []byte, keyLen uint32) []byte {
	return argon2.IDKey(password, salt, 2, 64*1024, 2, keyLen)
}
