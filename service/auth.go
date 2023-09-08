// Package service contains common functions used by
// the whole application
package service

import (
	"encoding/hex"

	"golang.org/x/crypto/blake2b"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
)

// GetUserByEmail ...
func GetUserByEmail(email string) (*model.Auth, error) {
	db := database.GetDB()

	var auth model.Auth

	if err := db.Where("email = ? ", email).First(&auth).Error; err != nil {
		return nil, err
	}

	return &auth, nil
}

// CalcEmailHash generates a fixed-sized BLAKE2b-256 hash of the email
func CalcEmailHash(email string, keyOptional []byte) (emailHash string, err error) {
	blake2b256Hash, err := blake2b.New256(keyOptional)
	if err != nil {
		return
	}

	_, err = blake2b256Hash.Write([]byte(email))
	if err != nil {
		return
	}

	blake2b256Sum := blake2b256Hash.Sum(nil)

	// hash of the email in hexadecimal string format
	emailHash = hex.EncodeToString(blake2b256Sum)

	return
}
