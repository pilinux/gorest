package service

import (
	"crypto/sha256"

	"github.com/pilinux/gorest/config"
)

// GetHash returns single or nested hash of the data
func GetHash(dataIn []byte) (dataOut []byte, err error) {
	hashed := sha256.Sum256(dataIn)

	if !config.Is2FADoubleHash() {
		dataOut = hashed[:]
		return
	}

	// second hashing
	dataOut, err = CalcHash(
		hashed[:],
		config.GetConfig().Security.Blake2bSec,
	)

	return
}
