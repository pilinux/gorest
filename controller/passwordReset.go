package controller

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mediocregopher/radix/v4"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"
)

// PasswordForgot sends secret code for resetting a forgotten password
func PasswordForgot(c *gin.Context) {
	payload := struct {
		Email string `json:"email"`
	}{}

	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	// check email format + perform mx lookup
	payload.Email = strings.TrimSpace(payload.Email)
	if !lib.ValidateEmail(payload.Email) {
		renderer.Render(c, gin.H{"msg": "wrong email address"}, http.StatusBadRequest)
		return
	}

	// find user
	v, err := service.GetUserByEmail(payload.Email)
	if err != nil {
		renderer.Render(c, gin.H{"msg": "user not found"}, http.StatusNotFound)
		return
	}

	// is email already verified
	if v.VerifyEmail != model.EmailVerified {
		renderer.Render(c, gin.H{"msg": "email not verified yet"}, http.StatusBadRequest)
		return
	}

	// send email with secret code
	if !sendEmail(v.Email, model.EmailTypePassRecovery) {
		renderer.Render(c, gin.H{"msg": "sending password recovery email not possible"}, http.StatusBadRequest)
		return
	}

	renderer.Render(c, gin.H{"msg": "sent password recovery email"}, http.StatusOK)
}

// PasswordRecover resets a forgotten password
func PasswordRecover(c *gin.Context) {
	// response to the client
	response := struct {
		Message     string `json:"msg,omitempty"`
		RecoveryKey string `json:"recoveryKey,omitempty"`
	}{}

	payload := struct {
		SecretCode  string `json:"secretCode"`
		PassNew     string `json:"passNew"`
		PassRepeat  string `json:"passRepeat"`
		RecoveryKey string `json:"recoveryKey"`
	}{}

	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	// app security settings
	configSecurity := config.GetConfig().Security

	// check minimum password length
	if len(payload.PassNew) < configSecurity.UserPassMinLength {
		msg := "password length must be greater than or equal to " + strconv.Itoa(configSecurity.UserPassMinLength)
		renderer.Render(c, gin.H{"msg": msg}, http.StatusBadRequest)
		return
	}

	// both passwords must be same
	if payload.PassNew != payload.PassRepeat {
		renderer.Render(c, gin.H{"msg": "password mismatch"}, http.StatusBadRequest)
		return
	}

	// for redis
	data := struct {
		key   string
		value string
	}{}
	data.key = model.PasswordRecoveryKeyPrefix + payload.SecretCode

	// get redis client
	client := *database.GetRedis()
	rConnTTL := config.GetConfig().Database.REDIS.Conn.ConnTTL

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// is key available in redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "EXISTS", data.key)); err != nil {
		log.WithError(err).Error("error code: 1081")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	if result == 0 {
		renderer.Render(c, gin.H{"msg": "wrong/expired secret code"}, http.StatusUnauthorized)
		return
	}

	// find key in redis
	if err := client.Do(ctx, radix.FlatCmd(&data.value, "GET", data.key)); err != nil {
		log.WithError(err).Error("error code: 1082")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
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
	pass, err := lib.HashPass(configHash, payload.PassNew)
	if err != nil {
		log.WithError(err).Error("error code: 1083")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	// get auth info from database
	db := database.GetDB()
	auth := model.Auth{}

	if err := db.Where("email = ?", data.value).First(&auth).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "unknown user"}, http.StatusUnauthorized)
		return
	}

	// current time
	timeNow := time.Now().Local()

	// is OTP required
	if configSecurity.Must2FA == config.Activated {
		payload.RecoveryKey = lib.RemoveAllSpace(payload.RecoveryKey)
		twoFA := model.TwoFA{}
		// is user account protected by 2FA
		if err := db.Where("id_auth = ?", auth.AuthID).First(&twoFA).Error; err == nil {
			if twoFA.Status == configSecurity.TwoFA.Status.On {
				// check recovery key length
				if len(payload.RecoveryKey) != configSecurity.TwoFA.Digits {
					renderer.Render(c, gin.H{"msg": "valid 2-fa recovery key required"}, http.StatusUnauthorized)
					return
				}

				// verify recovery key
				// step 1: hash recovery key in sha256
				hashRecoveryKey := sha256.Sum256([]byte(payload.RecoveryKey))

				// step 2: decode base64 encoded AES-256 encrypted uuid secret
				uuidCipherByte, err := base64.StdEncoding.DecodeString(twoFA.UUIDEnc)
				if err != nil {
					log.WithError(err).Error("error code: 1091")
					renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
					return
				}

				// step 3: decrypt (AES-256) uuid secret using hash of given recovery key
				// first verification: signature will fail for wrong recovery key
				uuidPlaintextByte, err := lib.Decrypt(uuidCipherByte, hashRecoveryKey[:])
				if err != nil {
					renderer.Render(c, gin.H{"msg": "invalid recovery key"}, http.StatusUnauthorized)
					return
				}
				// hash of decrypted uuid secret
				uuidPlaintextSHA := sha256.Sum256(uuidPlaintextByte)
				// second verification: compare
				uuidPlaintextBase64 := base64.StdEncoding.EncodeToString(uuidPlaintextSHA[:])
				if uuidPlaintextBase64 != twoFA.UUIDSHA {
					renderer.Render(c, gin.H{"msg": "invalid recovery key"}, http.StatusUnauthorized)
					return
				}
				// at this point, verification passed
				// now start process of encrypting existing keys with new pass and new recovery key

				// step 4: decode base64 encoded backup key
				keyBackupCipherByte, err := base64.StdEncoding.DecodeString(twoFA.KeyBackup)
				if err != nil {
					log.WithError(err).Error("error code: 1092")
					renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
					return
				}

				// step 5: decrypt (AES-256) backup key with hash of given recovery key
				keyBackupPlaintextByte, err := lib.Decrypt(keyBackupCipherByte, hashRecoveryKey[:])
				if err != nil {
					log.WithError(err).Error("error code: 1093")
					renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
					return
				}

				// step 6: generate new recovery key
				keyRecovery := uuid.NewString()
				keyRecovery = strings.ReplaceAll(keyRecovery, "-", "")
				keyRecovery = keyRecovery[len(keyRecovery)-configSecurity.TwoFA.Digits:]
				keyRecoveryHash := sha256.Sum256([]byte(keyRecovery))

				// step 7: encrypt secret (or backup key) with hash of new recovery key
				keyBackupCipherByte, err = lib.Encrypt(keyBackupPlaintextByte, keyRecoveryHash[:])
				if err != nil {
					log.WithError(err).Error("error code: 1094")
					renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
					return
				}

				// step 8: encrypt (AES-256) secret using hash of user's new pass
				passSHA := sha256.Sum256([]byte(payload.PassNew))
				keyMainCipherByte, err := lib.Encrypt(keyBackupPlaintextByte, passSHA[:])
				if err != nil {
					log.WithError(err).Error("error code: 1095")
					renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
					return
				}

				// step 9: generate new UUID code
				uuidPlaintext := uuid.NewString()
				uuidPlaintextByte = []byte(uuidPlaintext)
				uuidSHA256 := sha256.Sum256(uuidPlaintextByte)
				uuidSHA := base64.StdEncoding.EncodeToString(uuidSHA256[:])

				uuidEncByte, err := lib.Encrypt(uuidPlaintextByte, keyRecoveryHash[:])
				if err != nil {
					log.WithError(err).Error("error code: 1096")
					renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
					return
				}

				// step 10: encode in base64
				twoFA.KeyMain = base64.StdEncoding.EncodeToString(keyMainCipherByte)
				twoFA.KeyBackup = base64.StdEncoding.EncodeToString(keyBackupCipherByte)
				twoFA.UUIDEnc = base64.StdEncoding.EncodeToString(uuidEncByte)

				// update DB
				twoFA.UpdatedAt = timeNow
				twoFA.UUIDSHA = uuidSHA

				tx := db.Begin()
				if err := tx.Save(&twoFA).Error; err != nil {
					tx.Rollback()
					log.WithError(err).Error("error code: 1097")
					renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
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
		log.WithError(err).Error("error code: 1084")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	tx.Commit()

	// delete key from redis
	result = 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "DEL", data.key)); err != nil {
		log.WithError(err).Error("error code: 1085")
	}
	if result == 0 {
		err := errors.New("failed to delete recovery key from redis")
		log.WithError(err).Error("error code: 1086")
	}

	response.Message = "password updated"
	renderer.Render(c, response, http.StatusOK)
}

// PasswordUpdate - change password in logged-in state
func PasswordUpdate(c *gin.Context) {
	// get claims
	claims := getClaims(c)

	// check user validity
	ok := validateUserID(claims.AuthID, claims.Email)
	if !ok {
		renderer.Render(c, gin.H{"msg": "access denied"}, http.StatusUnauthorized)
		return
	}

	payload := struct {
		PassCurrent string `json:"passCurrent"`
		PassNew     string `json:"passNew"`
		PassRepeat  string `json:"passRepeat"`
	}{}

	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	// app security settings
	configSecurity := config.GetConfig().Security

	// check minimum password length
	if len(payload.PassNew) < configSecurity.UserPassMinLength {
		msg := "password length must be greater than or equal to " + strconv.Itoa(configSecurity.UserPassMinLength)
		renderer.Render(c, gin.H{"msg": msg}, http.StatusBadRequest)
		return
	}

	// both passwords must be same
	if payload.PassNew != payload.PassRepeat {
		renderer.Render(c, gin.H{"msg": "password mismatch"}, http.StatusBadRequest)
		return
	}

	// read DB
	db := database.GetDB()
	auth := model.Auth{}
	twoFA := model.TwoFA{}
	process2FA := false

	// auth info
	if err := db.Where("auth_id = ?", claims.AuthID).First(&auth).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "unknown user"}, http.StatusUnauthorized)
		return
	}

	// verify given pass against pass saved in DB
	verifyPass, err := argon2id.ComparePasswordAndHash(payload.PassCurrent, auth.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1091")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	if !verifyPass {
		renderer.Render(c, gin.H{"msg": "wrong credentials"}, http.StatusUnauthorized)
		return
	}

	// 2-FA info
	if configSecurity.Must2FA == config.Activated {
		if err := db.Where("id_auth = ?", claims.AuthID).First(&twoFA).Error; err == nil {
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
	pass, err := lib.HashPass(configHash, payload.PassNew)
	if err != nil {
		log.WithError(err).Error("error code: 1092")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	auth.Password = pass

	// current time
	timeNow := time.Now().Local()

	// process 2-FA
	if process2FA {
		// step 1: hash current password in sha256
		hashPassCurrent := sha256.Sum256([]byte(payload.PassCurrent))

		// step 2: hash new password in sha256
		hashPassNew := sha256.Sum256([]byte(payload.PassNew))

		// step 3: decode base64 encoded main key
		keyMainCipherByte, err := base64.StdEncoding.DecodeString(twoFA.KeyMain)
		if err != nil {
			log.WithError(err).Error("error code: 1093")
			renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
			return
		}

		// step 4: decrypt (AES-256) main key with hash of current password
		keyMainPlaintextByte, err := lib.Decrypt(keyMainCipherByte, hashPassCurrent[:])
		if err != nil {
			log.WithError(err).Error("error code: 1094")
			renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
			return
		}

		// step 5: encrypt secret (or main key) with hash of new password
		keyMainCipherByte, err = lib.Encrypt(keyMainPlaintextByte, hashPassNew[:])
		if err != nil {
			log.WithError(err).Error("error code: 1095")
			renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
			return
		}

		// step 6: encode in base64
		twoFA.KeyMain = base64.StdEncoding.EncodeToString(keyMainCipherByte)

		// update DB
		twoFA.UpdatedAt = timeNow
		tx := db.Begin()
		if err := tx.Save(&twoFA).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1096")
			renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
			return
		}
		tx.Commit()
	}

	auth.UpdatedAt = timeNow
	tx := db.Begin()
	if err := tx.Save(&auth).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1097")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	tx.Commit()

	renderer.Render(c, gin.H{"msg": "password updated"}, http.StatusOK)
}
