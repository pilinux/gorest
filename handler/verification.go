package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/mediocregopher/radix/v4"
	"github.com/pilinux/argon2"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
	"github.com/pilinux/gorest/service"
)

// VerifyEmail handles jobs for controller.VerifyEmail
func VerifyEmail(payload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	payload.VerificationCode = strings.TrimSpace(payload.VerificationCode)
	if payload.VerificationCode == "" {
		httpResponse.Message = "required a valid email verification code"
		httpStatusCode = http.StatusBadRequest
		return
	}

	data := struct {
		key   string
		value string
	}{}
	data.key = model.EmailVerificationKeyPrefix + payload.VerificationCode

	// get redis client
	client := *database.GetRedis()
	rConnTTL := config.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// is key available in redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "EXISTS", data.key)); err != nil {
		log.WithError(err).Error("error code: 1061.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if result == 0 {
		httpResponse.Message = "wrong/expired verification code"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// find key in redis
	if err := client.Do(ctx, radix.FlatCmd(&data.value, "GET", data.key)); err != nil {
		log.WithError(err).Error("error code: 1061.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// delete key from redis
	result = 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "DEL", data.key)); err != nil {
		log.WithError(err).Error("error code: 1061.3")
	}
	if result == 0 {
		err := errors.New("failed to delete recovery key from redis")
		log.WithError(err).Error("error code: 1061.4")
	}

	// update verification status in database
	db := database.GetDB()
	auth := model.Auth{}

	// is data.value an email or hash of an email
	isEmail := false
	if lib.ValidateEmail(data.value) {
		isEmail = true
	}

	if isEmail {
		if err := db.Where("email = ?", data.value).First(&auth).Error; err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1061.5")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}

			// email was in redis but not in relational db => missing data
			log.WithError(err).Error("error code: 1061.6")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}
	if !isEmail {
		if err := db.Where("email_hash = ?", data.value).First(&auth).Error; err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1061.7")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}

			// email hash was in redis but not in relational db => missing data
			log.WithError(err).Error("error code: 1061.8")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}

	if auth.VerifyEmail == model.EmailVerified {
		httpResponse.Message = "email already verified"
		httpStatusCode = http.StatusOK
		return
	}

	auth.VerifyEmail = model.EmailVerified
	auth.UpdatedAt = time.Now()

	tx := db.Begin()
	if err := tx.Save(&auth).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1061.9")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	httpResponse.Message = "email successfully verified"
	httpStatusCode = http.StatusOK
	return
}

// CreateVerificationEmail handles jobs for controller.CreateVerificationEmail
func CreateVerificationEmail(payload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	payload.Email = strings.TrimSpace(payload.Email)
	if !lib.ValidateEmail(payload.Email) {
		httpResponse.Message = "wrong email address"
		httpStatusCode = http.StatusBadRequest
		return
	}

	v, err := service.GetUserByEmail(payload.Email, true)
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1062.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		httpResponse.Message = "user not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	// is email already verified
	if v.VerifyEmail == model.EmailVerified {
		httpResponse.Message = "email already verified"
		httpStatusCode = http.StatusOK
		return
	}

	// verify password
	verifyPass, err := argon2.ComparePasswordAndHash(payload.Password, config.GetConfig().Security.HashSec, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1062.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !verifyPass {
		httpResponse.Message = "wrong credentials"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// issue new verification code
	emailDelivered, err := service.SendEmail(v.Email, model.EmailTypeVerifyEmailNewAcc)
	if err != nil {
		log.WithError(err).Error("error code: 1062.3")
		httpResponse.Message = "email delivery service failed"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !emailDelivered {
		httpResponse.Message = "failed to send verification email"
		httpStatusCode = http.StatusServiceUnavailable
		return
	}

	httpResponse.Message = "sent verification email"
	httpStatusCode = http.StatusOK
	return
}

// VerifyUpdatedEmail receives tasks from controller.VerifyUpdatedEmail
//
// - verify newly added email address
//
// - update user email address
//
// - delete temporary data from database after verification process is done
func VerifyUpdatedEmail(payload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	payload.VerificationCode = strings.TrimSpace(payload.VerificationCode)
	if payload.VerificationCode == "" {
		httpResponse.Message = "required a valid email verification code"
		httpStatusCode = http.StatusBadRequest
		return
	}

	data := struct {
		key   string
		value string
	}{}
	data.key = model.EmailUpdateKeyPrefix + payload.VerificationCode

	// get redis client
	client := *database.GetRedis()
	rConnTTL := config.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// is key available in redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "EXISTS", data.key)); err != nil {
		log.WithError(err).Error("error code: 1063.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if result == 0 {
		httpResponse.Message = "wrong/expired verification code"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// find key in redis
	if err := client.Do(ctx, radix.FlatCmd(&data.value, "GET", data.key)); err != nil {
		log.WithError(err).Error("error code: 1063.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// delete key from redis
	result = 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "DEL", data.key)); err != nil {
		log.WithError(err).Error("error code: 1063.3")
	}
	if result == 0 {
		err := errors.New("failed to delete recovery key from redis")
		log.WithError(err).Error("error code: 1063.4")
	}

	// update user email in database
	db := database.GetDB()
	auth := model.Auth{}
	tempEmail := model.TempEmail{}

	// is data.value an email or hash of an email
	isEmail := false
	if lib.ValidateEmail(data.value) {
		isEmail = true
	}

	if isEmail {
		// check 'temp_emails' with the email in plaintext
		if err := db.Where("email = ?", data.value).First(&tempEmail).Error; err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1063.5")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}

			// data was in redis but not in relational db
			// most probably using another verification code, the email was updated
			// hence, expire this request
			httpResponse.Message = "wrong/expired verification code"
			httpStatusCode = http.StatusBadRequest
			return
		}
	}
	if !isEmail {
		// check 'temp_emails' with hash of the email
		if err := db.Where("email_hash = ?", data.value).First(&tempEmail).Error; err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1063.6")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}

			// data was in redis but not in relational db
			// most probably using another verification code, the email was updated
			// hence, expire this request
			httpResponse.Message = "wrong/expired verification code"
			httpStatusCode = http.StatusBadRequest
			return
		}
	}

	// check 'auths'
	// verify that this email is not used by another user
	if isEmail {
		err := db.Where("email = ?", tempEmail.Email).First(&auth).Error

		if err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1063.71")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}
		}
		if err == nil {
			httpResponse.Message = "email already in use"
			httpStatusCode = http.StatusBadRequest
			return
		}
	}
	if !isEmail {
		err := db.Where("email_hash = ?", tempEmail.EmailHash).First(&auth).Error

		if err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1063.72")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}
		}
		if err == nil {
			httpResponse.Message = "email already in use"
			httpStatusCode = http.StatusBadRequest
			return
		}
	}

	// fetch auth data
	if err := db.Where("auth_id = ?", tempEmail.IDAuth).First(&auth).Error; err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1063.73")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// auth_id mismatch!
		httpResponse.Message = "wrong/expired verification code"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// update model
	timeNow := time.Now()
	auth.UpdatedAt = timeNow
	auth.Email = tempEmail.Email
	auth.EmailCipher = tempEmail.EmailCipher
	auth.EmailNonce = tempEmail.EmailNonce
	auth.EmailHash = tempEmail.EmailHash

	// delete data from 'temp_emails'
	tx := db.Begin()
	if err := tx.Delete(&tempEmail).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1063.8")
	} else {
		tx.Commit()
	}

	// update data in 'auths'
	tx = db.Begin()
	if err := tx.Save(&auth).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1063.9")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	httpResponse.Message = "email address updated"
	httpStatusCode = http.StatusOK
	return
}

// GetUnverifiedEmail receives tasks from controller.GetUnverifiedEmail
//
// It retrieves unverified email information for a given user.
func GetUnverifiedEmail(claims middleware.MyCustomClaims) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// check auth validity
	ok := service.ValidateAuthID(claims.AuthID)
	if !ok {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// read DB
	db := database.GetDB()
	tempEmail := model.TempEmail{}

	// check 'temp_emails'
	err := db.Where("id_auth = ?", claims.AuthID).First(&tempEmail).Error
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1064.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		httpResponse.Message = "no pending request"
	}

	// verification is pending to modify current email
	if err == nil {
		// decipher
		if tempEmail.Email == "" {
			if !config.IsCipher() {
				e := errors.New("check env: ACTIVATE_CIPHER")
				log.WithError(e).Error("error code: 1064.2")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}

			tempEmail.Email, err = service.DecryptEmail(tempEmail.EmailNonce, tempEmail.EmailCipher)
			if err != nil {
				log.WithError(err).Error("error code: 1064.3")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}
		}

		// clear cipher data
		tempEmail.EmailCipher = ""
		tempEmail.EmailNonce = ""
		tempEmail.EmailHash = ""

		httpResponse.Message = tempEmail
	}

	httpStatusCode = http.StatusOK
	return
}

// ResendVerificationCodeToModifyActiveEmail receives tasks from controller.ResendVerificationCodeToModifyActiveEmail
func ResendVerificationCodeToModifyActiveEmail(claims middleware.MyCustomClaims) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// check auth validity
	ok := service.ValidateAuthID(claims.AuthID)
	if !ok {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// read DB
	db := database.GetDB()
	tempEmail := model.TempEmail{}

	// check 'temp_emails'
	err := db.Where("id_auth = ?", claims.AuthID).First(&tempEmail).Error
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1065.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		httpResponse.Message = "no pending request"
		httpStatusCode = http.StatusBadRequest
		return
	}
	// verification is pending to modify current email
	// decipher
	if tempEmail.Email == "" {
		if !config.IsCipher() {
			e := errors.New("check env: ACTIVATE_CIPHER")
			log.WithError(e).Error("error code: 1065.2")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		tempEmail.Email, err = service.DecryptEmail(tempEmail.EmailNonce, tempEmail.EmailCipher)
		if err != nil {
			log.WithError(err).Error("error code: 1065.3")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}

	// issue new verification code
	emailDelivered, err := service.SendEmail(tempEmail.Email, model.EmailTypeVerifyUpdatedEmail)
	if err != nil {
		log.WithError(err).Error("error code: 1065.4")
		httpResponse.Message = "email delivery service failed"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !emailDelivered {
		httpResponse.Message = "failed to send verification email"
		httpStatusCode = http.StatusServiceUnavailable
		return
	}

	httpResponse.Message = "sent verification email"
	httpStatusCode = http.StatusOK
	return
}
