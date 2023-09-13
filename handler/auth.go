// Package handler sits in between controller and database services.
package handler

import (
	"encoding/hex"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/pilinux/crypt"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/service"
)

// CreateUserAuth receives tasks from controller.CreateUserAuth.
// After email validation, it creates a new user account. It
// supports both the legacy way of saving user email in plaintext
// and the recommended way of applying encryption at rest.
func CreateUserAuth(auth model.Auth) (httpResponse model.HTTPResponse, httpStatusCode int) {
	db := database.GetDB()

	// user must not be able to manipulate all fields
	authFinal := new(model.Auth)
	authFinal.Email = auth.Email
	authFinal.Password = auth.Password

	// email validation
	if !lib.ValidateEmail(auth.Email) {
		httpResponse.Message = "wrong email address"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// for backward compatibility
	// email must be unique
	err := db.Where("email = ?", auth.Email).First(&auth).Error
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1002.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}
	if err == nil {
		httpResponse.Message = "email already registered"
		httpStatusCode = http.StatusForbidden
		return
	}

	// downgrade must be avoided to prevent creating duplicate accounts
	// valid: non-encryption mode -> upgrade to encryption mode
	// invalid: encryption mode -> downgrade to non-encryption mode
	if !config.IsCipher() {
		err := db.Where("email_hash IS NOT NULL AND email_hash != ?", "").First(&auth).Error
		if err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1002.2")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}
		}
		if err == nil {
			log.Error("check env: ACTIVATE_CIPHER")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}

	// generate a fixed-sized BLAKE2b-256 hash of the email, used for auth purpose
	// when encryption at rest is used
	if config.IsCipher() {
		var err error

		// hash of the email in hexadecimal string format
		emailHash, err := service.CalcHash(
			[]byte(auth.Email),
			config.GetConfig().Security.Blake2bSec,
		)
		if err != nil {
			log.WithError(err).Error("error code: 1001.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		authFinal.EmailHash = hex.EncodeToString(emailHash)

		// email must be unique
		err = db.Where("email_hash = ?", authFinal.EmailHash).First(&auth).Error
		if err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1002.3")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}
		}
		if err == nil {
			httpResponse.Message = "email already registered"
			httpStatusCode = http.StatusForbidden
			return
		}
	}

	// send a verification email if required by the application
	emailDelivered, err := service.SendEmail(authFinal.Email, model.EmailTypeVerification)
	if err != nil {
		log.WithError(err).Error("error code: 1002.4")
		httpResponse.Message = "email delivery service failed"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if err == nil {
		if emailDelivered {
			authFinal.VerifyEmail = model.EmailNotVerified
		}
	}

	// encryption at rest for user email, mainly needed by system in future
	// to send verification or password recovery emails
	if config.IsCipher() {
		// encrypt the email
		cipherEmail, nonce, err := crypt.EncryptChacha20poly1305(
			config.GetConfig().Security.CipherKey,
			auth.Email,
		)
		if err != nil {
			log.WithError(err).Error("error code: 1001.2")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// save email only in ciphertext
		authFinal.Email = ""
		authFinal.EmailCipher = hex.EncodeToString(cipherEmail)
		authFinal.EmailNonce = hex.EncodeToString(nonce)
	}

	// one unique email for each account
	tx := db.Begin()
	if err := tx.Create(&authFinal).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1001.3")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	httpResponse.Message = *authFinal
	httpStatusCode = http.StatusCreated
	return
}
