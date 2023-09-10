package service

import (
	"crypto/rand"
	"crypto/sha256"
	"math/big"

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

// RandomByte returns a secure random byte slice of the given length
func RandomByte(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateCode generates a random alphanumeric code of the given length
func GenerateCode(length int) (string, error) {
	const characters string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	var code []byte
	maxIndex := big.NewInt(int64(len(characters)))

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, maxIndex)
		if err != nil {
			return "", err
		}
		code = append(code, characters[randomIndex.Int64()])
	}

	return string(code), nil
}
