// Package handler sits in between controller and database services.
package handler

import (
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/pilinux/argon2"
	"github.com/pilinux/crypt"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
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
		httpStatusCode = http.StatusBadRequest
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
			e := errors.New("check env: ACTIVATE_CIPHER")
			log.WithError(e).Error("error code: 1002.3")
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
				log.WithError(err).Error("error code: 1002.4")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}
		}
		if err == nil {
			httpResponse.Message = "email already registered"
			httpStatusCode = http.StatusBadRequest
			return
		}
	}

	// send a verification email if required by the application
	emailDelivered, err := service.SendEmail(authFinal.Email, model.EmailTypeVerifyEmailNewAcc)
	if err != nil {
		log.WithError(err).Error("error code: 1002.5")
		httpResponse.Message = "email delivery service failed"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if emailDelivered {
		authFinal.VerifyEmail = model.EmailNotVerified
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

// UpdateEmail receives tasks from controller.UpdateEmail.
//
// step 1: validate email format
//
// step 2: verify that this email is not registered to anyone
//
// step 3: load user credentials
//
// step 4: verify user password
//
// step 5: calculate hash of the new email
//
// step 6: read 'temp_emails' table
//
// step 7: verify that this is not a repeated request for the same email
//
// step 8: populate model with data to be processed in database
//
// step 9: send a verification email if required by the app
func UpdateEmail(claims middleware.MyCustomClaims, req model.TempEmail) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// check auth validity
	ok := service.ValidateAuthID(claims.AuthID)
	if !ok {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// step 1: validate email format
	req.Email = strings.TrimSpace(req.Email)
	if !lib.ValidateEmail(req.Email) {
		httpResponse.Message = "wrong email address"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// step 2: verify that this email is not registered to anyone
	_, err := service.GetUserByEmail(req.Email, false)
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1003.21")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}
	if err == nil {
		httpResponse.Message = "email already registered"
		httpStatusCode = http.StatusBadRequest
		return
	}
	// ok: email is not registered yet, continue...

	// db connection
	db := database.GetDB()

	// step 3: load user credentials
	auth := model.Auth{}
	err = db.Where("auth_id = ?", claims.AuthID).First(&auth).Error
	if err != nil {
		// most likely db read error
		log.WithError(err).Error("error code: 1003.31")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// app settings
	configSecurity := config.GetConfig().Security

	// step 4: verify user password
	verifyPass, err := argon2.ComparePasswordAndHash(req.Password, configSecurity.HashSec, auth.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1003.41")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !verifyPass {
		httpResponse.Message = "wrong password"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// step 5: calculate hash of the new email
	emailHash, err := service.CalcHash(
		[]byte(req.Email),
		config.GetConfig().Security.Blake2bSec,
	)
	if err != nil {
		log.WithError(err).Error("error code: 1003.51")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 6: read 'temp_emails' table
	tEmailDB := model.TempEmail{}
	err = db.Where("id_auth = ?", claims.AuthID).First(&tEmailDB).Error
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1003.61")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		// this user has no previous pending request to update email
	}

	// this user has a pending request to update the email
	if err == nil {
		// step 7: verify that this is not a repeated request for the same email

		// plaintext
		if tEmailDB.Email != "" {
			if tEmailDB.Email == req.Email {
				httpResponse.Message = "please verify the new email"
				httpStatusCode = http.StatusBadRequest
				return
			}
		}

		// encryption at rest
		if tEmailDB.Email == "" {
			if tEmailDB.EmailHash == hex.EncodeToString(emailHash) {
				httpResponse.Message = "please verify the new email"
				httpStatusCode = http.StatusBadRequest
				return
			}
		}
	}

	// step 8: populate model with data to be processed in database
	timeNow := time.Now()

	// create new data
	if tEmailDB.ID == 0 {
		tEmailDB.CreatedAt = timeNow
		tEmailDB.IDAuth = claims.AuthID
	}

	tEmailDB.UpdatedAt = timeNow

	// plaintext
	if !config.IsCipher() {
		tEmailDB.Email = req.Email
		tEmailDB.EmailCipher = ""
		tEmailDB.EmailNonce = ""
		tEmailDB.EmailHash = ""
	}

	// encryption at rest
	if config.IsCipher() {
		tEmailDB.Email = ""

		// encrypt the email
		cipherEmail, nonce, err := crypt.EncryptChacha20poly1305(
			config.GetConfig().Security.CipherKey,
			req.Email,
		)
		if err != nil {
			log.WithError(err).Error("error code: 1003.81")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// save email only in ciphertext
		tEmailDB.EmailCipher = hex.EncodeToString(cipherEmail)
		tEmailDB.EmailNonce = hex.EncodeToString(nonce)
		tEmailDB.EmailHash = hex.EncodeToString(emailHash)
	}

	// step 9: send a verification email if required by the application
	emailDelivered, err := service.SendEmail(req.Email, model.EmailTypeVerifyUpdatedEmail)
	if err != nil {
		log.WithError(err).Error("error code: 1003.91")
		httpResponse.Message = "email delivery service failed"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// verification code sent
	if emailDelivered {
		tx := db.Begin()
		if err := tx.Save(&tEmailDB).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1003.92")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		tx.Commit()

		httpResponse.Message = "verification email delivered"
		httpStatusCode = http.StatusOK
	}

	// verification code not required, update email immediately
	if !emailDelivered {
		auth.UpdatedAt = timeNow
		auth.Email = tEmailDB.Email
		auth.EmailCipher = tEmailDB.EmailCipher
		auth.EmailNonce = tEmailDB.EmailNonce
		auth.EmailHash = tEmailDB.EmailHash

		tx := db.Begin()
		if err := tx.Save(&auth).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1003.93")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		tx.Commit()

		httpResponse.Message = "email updated"
		httpStatusCode = http.StatusOK
	}

	return
}
