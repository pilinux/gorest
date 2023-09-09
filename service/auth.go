// Package service contains common functions used by
// the whole application
package service

import (
	"encoding/hex"

	"github.com/pilinux/crypt"
	"golang.org/x/crypto/blake2b"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
)

// GetUserByEmail fetches auth info by email or hash of the email
func GetUserByEmail(email string, decryptEmail bool) (*model.Auth, error) {
	db := database.GetDB()
	var err error

	var auth model.Auth

	// when email is saved in plaintext
	if err = db.Where("email = ? ", email).First(&auth).Error; err == nil {
		return &auth, nil
	}

	// encryption at rest
	if config.IsCipher() {
		// hash of the email in hexadecimal string format
		emailHash, err := CalcHash(
			email,
			config.GetConfig().Security.Blake2bSec,
		)
		if err != nil {
			return nil, err
		}

		// email must be unique
		if err = db.Where("email_hash = ?", emailHash).First(&auth).Error; err == nil {
			if decryptEmail {
				nonce, err := hex.DecodeString(auth.EmailNonce)
				if err != nil {
					return nil, err
				}
				cipherEmail, err := hex.DecodeString(auth.EmailCipher)
				if err != nil {
					return nil, err
				}

				auth.Email, err = crypt.DecryptChacha20poly1305(
					config.GetConfig().Security.CipherKey,
					nonce,
					cipherEmail,
				)
				if err != nil {
					return nil, err
				}
			}

			return &auth, nil
		}
	}

	return nil, err
}

// CalcHash generates a fixed-sized BLAKE2b-256 hash of the given text
func CalcHash(plaintext string, keyOptional []byte) (hashedText string, err error) {
	blake2b256Hash, err := blake2b.New256(keyOptional)
	if err != nil {
		return
	}

	_, err = blake2b256Hash.Write([]byte(plaintext))
	if err != nil {
		return
	}

	blake2b256Sum := blake2b256Hash.Sum(nil)

	// hash of the email in hexadecimal string format
	hashedText = hex.EncodeToString(blake2b256Sum)

	return
}
