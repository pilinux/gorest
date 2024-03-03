package handler

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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

// PasswordForgot handles jobs for controller.PasswordForgot
func PasswordForgot(authPayload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// check email format + perform mx lookup
	authPayload.Email = strings.TrimSpace(authPayload.Email)
	if !lib.ValidateEmail(authPayload.Email) {
		httpResponse.Message = "wrong email address"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// find user
	v, err := service.GetUserByEmail(authPayload.Email, true)
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1030.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		httpResponse.Message = "user not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	// is email already verified
	if v.VerifyEmail != model.EmailVerified {
		httpResponse.Message = "email not verified yet"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// send email with secret code
	emailDelivered, err := service.SendEmail(v.Email, model.EmailTypePassRecovery)
	if err != nil {
		log.WithError(err).Error("error code: 1030.2")
		httpResponse.Message = "email delivery service failed"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !emailDelivered {
		httpResponse.Message = "sending password recovery email not possible"
		httpStatusCode = http.StatusServiceUnavailable
		return
	}

	httpResponse.Message = "sent password recovery email"
	httpStatusCode = http.StatusOK
	return
}

// PasswordRecover handles jobs for controller.PasswordRecover
func PasswordRecover(authPayload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// response to the client
	response := struct {
		Message     string `json:"message,omitempty"`
		RecoveryKey string `json:"recoveryKey,omitempty"`
	}{}

	// app security settings
	configSecurity := config.GetConfig().Security

	// check minimum password length
	if len(authPayload.PassNew) < configSecurity.UserPassMinLength {
		msg := "password length must be greater than or equal to " + strconv.Itoa(configSecurity.UserPassMinLength)
		httpResponse.Message = msg
		httpStatusCode = http.StatusBadRequest
		return
	}

	// both passwords must be same
	if authPayload.PassNew != authPayload.PassRepeat {
		httpResponse.Message = "password mismatch"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// for redis
	data := struct {
		key   string
		value string
	}{}
	data.key = model.PasswordRecoveryKeyPrefix + authPayload.SecretCode

	// get redis client
	client := *database.GetRedis()
	rConnTTL := config.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// is key available in redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "EXISTS", data.key)); err != nil {
		log.WithError(err).Error("error code: 1021.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if result == 0 {
		httpResponse.Message = "wrong/expired secret code"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// find key in redis
	if err := client.Do(ctx, radix.FlatCmd(&data.value, "GET", data.key)); err != nil {
		log.WithError(err).Error("error code: 1021.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// hashing
	configHash := lib.HashPassConfig{
		Memory:      configSecurity.HashPass.Memory,
		Iterations:  configSecurity.HashPass.Iterations,
		Parallelism: configSecurity.HashPass.Parallelism,
		SaltLength:  configSecurity.HashPass.SaltLength,
		KeyLength:   configSecurity.HashPass.KeyLength,
	}
	pass, err := lib.HashPass(configHash, authPayload.PassNew, configSecurity.HashSec)
	if err != nil {
		log.WithError(err).Error("error code: 1021.3")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// get auth info from database
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
				log.WithError(err).Error("error code: 1021.4")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}

			// most likely system admin manually deleted this account?
			httpResponse.Message = "unknown user"
			httpStatusCode = http.StatusBadRequest
			return
		}
	}
	if !isEmail {
		if err := db.Where("email_hash = ?", data.value).First(&auth).Error; err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1021.5")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}

			// most likely system admin manually deleted this account?
			httpResponse.Message = "unknown user"
			httpStatusCode = http.StatusBadRequest
			return
		}
	}

	// current time
	timeNow := time.Now()

	// is OTP required
	if configSecurity.Must2FA == config.Activated {
		authPayload.RecoveryKey = lib.RemoveAllSpace(authPayload.RecoveryKey)
		twoFA := model.TwoFA{}
		// is user account protected by 2FA
		err := db.Where("id_auth = ?", auth.AuthID).First(&twoFA).Error
		if err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1021.6")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}
		}
		if err == nil {
			if twoFA.Status == configSecurity.TwoFA.Status.On {
				// check recovery key length
				if len(authPayload.RecoveryKey) != configSecurity.TwoFA.Digits {
					httpResponse.Message = "valid 2-fa recovery key required"
					httpStatusCode = http.StatusUnauthorized
					return
				}

				// verify recovery key
				// step 1: hash recovery key
				hashRecoveryKey, err := service.GetHash([]byte(authPayload.RecoveryKey))
				if err != nil {
					log.WithError(err).Error("error code: 1022.1")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}

				// step 2: decode base64 encoded AES-256 encrypted uuid secret
				uuidCipherByte, err := base64.StdEncoding.DecodeString(twoFA.UUIDEnc)
				if err != nil {
					log.WithError(err).Error("error code: 1022.2")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}

				// step 3: decrypt (AES-256) uuid secret using hash of given recovery key
				// first verification: signature will fail for wrong recovery key
				uuidPlaintextByte, err := lib.Decrypt(uuidCipherByte, hashRecoveryKey)
				if err != nil {
					httpResponse.Message = "invalid recovery key"
					httpStatusCode = http.StatusUnauthorized
					return
				}
				// hash of decrypted uuid secret
				uuidPlaintextSHA, err := service.GetHash(uuidPlaintextByte)
				if err != nil {
					log.WithError(err).Error("error code: 1022.3")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}
				// second verification: compare
				uuidPlaintextBase64 := base64.StdEncoding.EncodeToString(uuidPlaintextSHA)
				if uuidPlaintextBase64 != twoFA.UUIDSHA {
					httpResponse.Message = "invalid recovery key"
					httpStatusCode = http.StatusUnauthorized
					return
				}
				// at this point, verification passed
				// now start process of encrypting existing keys with new pass and new recovery key

				// step 4: decode base64 encoded backup key
				keyBackupCipherByte, err := base64.StdEncoding.DecodeString(twoFA.KeyBackup)
				if err != nil {
					log.WithError(err).Error("error code: 1023.1")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}

				// step 5: decrypt (AES-256) backup key with hash of given recovery key
				keyBackupPlaintextByte, err := lib.Decrypt(keyBackupCipherByte, hashRecoveryKey)
				if err != nil {
					log.WithError(err).Error("error code: 1023.2")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}

				// step 6: generate new recovery key
				keyRecovery := uuid.NewString()
				keyRecovery = strings.ReplaceAll(keyRecovery, "-", "")
				keyRecovery = keyRecovery[len(keyRecovery)-configSecurity.TwoFA.Digits:]
				keyRecoveryHash, err := service.GetHash([]byte(keyRecovery))
				if err != nil {
					log.WithError(err).Error("error code: 1024.1")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}

				// step 7: encrypt secret (or backup key) with hash of new recovery key
				keyBackupCipherByte, err = lib.Encrypt(keyBackupPlaintextByte, keyRecoveryHash)
				if err != nil {
					log.WithError(err).Error("error code: 1024.2")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}

				// step 8: encrypt (AES-256) secret using hash of user's new pass
				passSHA, err := service.GetHash([]byte(authPayload.PassNew))
				if err != nil {
					log.WithError(err).Error("error code: 1024.3")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}
				keyMainCipherByte, err := lib.Encrypt(keyBackupPlaintextByte, passSHA)
				if err != nil {
					log.WithError(err).Error("error code: 1024.4")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}

				// step 9: generate new UUID code
				uuidPlaintext := uuid.NewString()
				uuidPlaintextByte = []byte(uuidPlaintext)
				uuidSHA, err := service.GetHash(uuidPlaintextByte)
				if err != nil {
					log.WithError(err).Error("error code: 1024.5")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}
				uuidSHAStr := base64.StdEncoding.EncodeToString(uuidSHA)

				uuidEncByte, err := lib.Encrypt(uuidPlaintextByte, keyRecoveryHash)
				if err != nil {
					log.WithError(err).Error("error code: 1024.6")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}

				// step 10: encode in base64
				twoFA.KeyMain = base64.StdEncoding.EncodeToString(keyMainCipherByte)
				twoFA.KeyBackup = base64.StdEncoding.EncodeToString(keyBackupCipherByte)
				twoFA.UUIDEnc = base64.StdEncoding.EncodeToString(uuidEncByte)

				// update DB
				twoFA.UpdatedAt = timeNow
				twoFA.UUIDSHA = uuidSHAStr

				tx := db.Begin()
				if err := tx.Save(&twoFA).Error; err != nil {
					tx.Rollback()
					log.WithError(err).Error("error code: 1024.7")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}
				tx.Commit()

				response.RecoveryKey = keyRecovery
			}
		}
	}

	auth.UpdatedAt = timeNow
	auth.Password = pass

	tx := db.Begin()
	if err := tx.Save(&auth).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1025.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	// delete key from redis
	result = 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "DEL", data.key)); err != nil {
		log.WithError(err).Error("error code: 1025.2")
	}
	if result == 0 {
		err := errors.New("failed to delete password recovery secret key from redis")
		log.WithError(err).Error("error code: 1025.3")
	}

	response.Message = "password updated"
	httpResponse.Message = response
	httpStatusCode = http.StatusOK
	return
}

// PasswordUpdate handles jobs for controller.PasswordUpdate
func PasswordUpdate(claims middleware.MyCustomClaims, authPayload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// check auth validity
	ok := service.ValidateAuthID(claims.AuthID)
	if !ok {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// app security settings
	configSecurity := config.GetConfig().Security

	// check minimum password length
	if len(authPayload.PassNew) < configSecurity.UserPassMinLength {
		msg := "password length must be greater than or equal to " + strconv.Itoa(configSecurity.UserPassMinLength)
		httpResponse.Message = msg
		httpStatusCode = http.StatusBadRequest
		return
	}

	// both passwords must be same
	if authPayload.PassNew != authPayload.PassRepeat {
		httpResponse.Message = "password mismatch"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// read DB
	db := database.GetDB()
	auth := model.Auth{}
	twoFA := model.TwoFA{}
	process2FA := false

	// auth info
	if err := db.Where("auth_id = ?", claims.AuthID).First(&auth).Error; err != nil {
		// most likely db read error
		log.WithError(err).Error("error code: 1026.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// verify given pass against pass saved in DB
	verifyPass, err := argon2.ComparePasswordAndHash(authPayload.Password, configSecurity.HashSec, auth.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1026.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !verifyPass {
		httpResponse.Message = "wrong credentials"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// 2-FA info
	if configSecurity.Must2FA == config.Activated {
		err := db.Where("id_auth = ?", claims.AuthID).First(&twoFA).Error
		if err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1026.3")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}
		}
		if err == nil {
			if twoFA.Status == configSecurity.TwoFA.Status.On {
				process2FA = true
			}
		}
	}

	// argon2id hashing of new password
	configHash := lib.HashPassConfig{
		Memory:      configSecurity.HashPass.Memory,
		Iterations:  configSecurity.HashPass.Iterations,
		Parallelism: configSecurity.HashPass.Parallelism,
		SaltLength:  configSecurity.HashPass.SaltLength,
		KeyLength:   configSecurity.HashPass.KeyLength,
	}
	pass, err := lib.HashPass(configHash, authPayload.PassNew, configSecurity.HashSec)
	if err != nil {
		log.WithError(err).Error("error code: 1026.4")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	auth.Password = pass

	// current time
	timeNow := time.Now()

	// process 2-FA
	if process2FA {
		// step 1: hash current password
		hashPassCurrent, err := service.GetHash([]byte(authPayload.Password))
		if err != nil {
			log.WithError(err).Error("error code: 1027.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// step 2: hash new password
		hashPassNew, err := service.GetHash([]byte(authPayload.PassNew))
		if err != nil {
			log.WithError(err).Error("error code: 1027.2")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// step 3: decode base64 encoded main key
		keyMainCipherByte, err := base64.StdEncoding.DecodeString(twoFA.KeyMain)
		if err != nil {
			log.WithError(err).Error("error code: 1027.3")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// step 4: decrypt (AES-256) main key with hash of current password
		keyMainPlaintextByte, err := lib.Decrypt(keyMainCipherByte, hashPassCurrent)
		if err != nil {
			log.WithError(err).Error("error code: 1028.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// step 5: encrypt secret (or main key) with hash of new password
		keyMainCipherByte, err = lib.Encrypt(keyMainPlaintextByte, hashPassNew)
		if err != nil {
			log.WithError(err).Error("error code: 1028.2")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// step 6: encode in base64
		twoFA.KeyMain = base64.StdEncoding.EncodeToString(keyMainCipherByte)

		// update DB
		twoFA.UpdatedAt = timeNow
		tx := db.Begin()
		if err := tx.Save(&twoFA).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1029.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		tx.Commit()
	}

	auth.UpdatedAt = timeNow
	tx := db.Begin()
	if err := tx.Save(&auth).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1029.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	httpResponse.Message = "password updated"
	httpStatusCode = http.StatusOK
	return
}
