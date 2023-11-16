package handler

import (
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pilinux/argon2"
	"github.com/pilinux/crypt"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
	"github.com/pilinux/gorest/service"
)

// Setup2FA handles jobs for controller.Setup2FA
func Setup2FA(claims middleware.MyCustomClaims, authPayload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// check auth validity
	ok := service.ValidateAuthID(claims.AuthID)
	if !ok {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// 2FA already enabled
	configSecurity := config.GetConfig().Security
	if claims.TwoFA == configSecurity.TwoFA.Status.Verified {
		// JWT: 2FA verified, abort setup
		httpResponse.Message = "twoFA: " + configSecurity.TwoFA.Status.Verified
		httpStatusCode = http.StatusOK
		return
	}
	if claims.TwoFA == configSecurity.TwoFA.Status.On {
		// JWT: 2FA ON, abort setup
		httpResponse.Message = "twoFA: " + configSecurity.TwoFA.Status.On
		httpStatusCode = http.StatusBadRequest
		return
	}

	// is 2FA disabled/never configured before
	db := database.GetDB()
	twoFA := model.TwoFA{}
	// err == RecordNotFound => never configured before
	// err == nil => 2FA disabled
	err := db.Where("id_auth = ?", claims.AuthID).First(&twoFA).Error
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1051.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}
	if err == nil {
		if twoFA.Status == configSecurity.TwoFA.Status.On {
			// DB: 2FA ON, abort setup
			httpResponse.Message = "twoFA: " + configSecurity.TwoFA.Status.On
			httpStatusCode = http.StatusBadRequest
			return
		}
	}

	// retrieve user email
	v := model.Auth{}
	err = db.Where("auth_id = ?", claims.AuthID).First(&v).Error
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1051.2")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		httpResponse.Message = "user not found"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// start OTP setup procedure
	//
	// step 1: verify user pass
	verifyPass, err := argon2.ComparePasswordAndHash(authPayload.Password, configSecurity.HashSec, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1051.11")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !verifyPass {
		httpResponse.Message = "wrong credentials"
		httpStatusCode = http.StatusBadRequest
		return
	}
	// get user email
	if !lib.ValidateEmail(v.Email) {
		nonce, err := hex.DecodeString(v.EmailNonce)
		if err != nil {
			log.WithError(err).Error("error code: 1051.12")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		cipherEmail, err := hex.DecodeString(v.EmailCipher)
		if err != nil {
			log.WithError(err).Error("error code: 1051.13")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		v.Email, err = crypt.DecryptChacha20poly1305(
			config.GetConfig().Security.CipherKey,
			nonce,
			cipherEmail,
		)
		if err != nil {
			log.WithError(err).Error("error code: 1051.14")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}

	// step 2: create new TOTP object
	otpByte, err := lib.NewTOTP(
		v.Email,
		configSecurity.TwoFA.Issuer,
		configSecurity.TwoFA.Crypto,
		configSecurity.TwoFA.Digits,
	)
	if err != nil {
		log.WithError(err).Error("error code: 1051.21")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 3: encode QR in bytes
	qrByte, err := lib.NewQR(
		otpByte,
		configSecurity.TwoFA.Issuer,
	)
	if err != nil {
		log.WithError(err).Error("error code: 1051.31")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 4: generate QR in PNG format and save on disk
	img, err := lib.ByteToPNG(qrByte, configSecurity.TwoFA.PathQR+"/")
	if err != nil {
		log.WithError(err).Error("error code: 1051.41")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 5: hash user's pass
	hashPass, err := service.GetHash([]byte(authPayload.Password))
	if err != nil {
		log.WithError(err).Error("error code: 1051.51")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 6: check if client secret is available in memory
	data2FA, ok := model.InMemorySecret2FA[claims.AuthID]
	if ok {
		// delete old QR image if available
		if lib.FileExist(configSecurity.TwoFA.PathQR + "/" + data2FA.Image) {
			err := os.Remove(configSecurity.TwoFA.PathQR + "/" + data2FA.Image)
			if err != nil {
				log.WithError(err).Error("error code: 1051.61")
			}
		}
	}

	// step 7: save the secrets in memory for validation
	data2FA.PassSHA = hashPass
	data2FA.Secret = otpByte
	data2FA.Image = img
	model.InMemorySecret2FA[claims.AuthID] = data2FA

	// serve the QR to the client
	httpResponse.Message = configSecurity.TwoFA.PathQR + "/" + img
	httpStatusCode = http.StatusCreated
	return
}

// Activate2FA handles jobs for controller.Activate2FA
func Activate2FA(claims middleware.MyCustomClaims, authPayload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// check auth validity
	ok := service.ValidateAuthID(claims.AuthID)
	if !ok {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	configSecurity := config.GetConfig().Security
	if claims.TwoFA == configSecurity.TwoFA.Status.Verified {
		// JWT: 2FA verified, abort setup
		httpResponse.Message = "twoFA: " + configSecurity.TwoFA.Status.Verified
		httpStatusCode = http.StatusBadRequest
		return
	}
	if claims.TwoFA == configSecurity.TwoFA.Status.On {
		// JWT: 2FA ON, abort setup
		httpResponse.Message = "twoFA: " + configSecurity.TwoFA.Status.On
		httpStatusCode = http.StatusBadRequest
		return
	}

	// start 2FA activation procedure
	//
	// step 1: check if client secret is available in memory
	data2FA, ok := model.InMemorySecret2FA[claims.AuthID]
	if !ok {
		// request user to visit setup endpoint first
		httpResponse.Message = "request for a new 2-fa secret"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// step 2: check otp length
	authPayload.OTP = lib.RemoveAllSpace(authPayload.OTP)
	if len(authPayload.OTP) != configSecurity.TwoFA.Digits {
		httpResponse.Message = "wrong one-time password"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// step 3: validate user-provided OTP
	otpByte, status, err := service.Validate2FA(
		data2FA.Secret,
		configSecurity.TwoFA.Issuer,
		authPayload.OTP,
	)
	if err != nil {
		// client provided invalid OTP
		if status == configSecurity.TwoFA.Status.Invalid {
			// save the secret with failed attempt in memory for future validation procedure
			data2FA.Secret = otpByte
			model.InMemorySecret2FA[claims.AuthID] = data2FA

			httpResponse.Message = "wrong one-time password"
			httpStatusCode = http.StatusBadRequest
			return
		}

		// internal error
		log.WithError(err).Error("error code: 1052.31")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 4: check DB
	db := database.GetDB()
	twoFA := model.TwoFA{}
	available := false
	err = db.Where("id_auth = ?", claims.AuthID).First(&twoFA).Error
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1052.41")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}
	if err == nil {
		// record found in DB
		available = true

		// 2FA already activated!
		if twoFA.Status == configSecurity.TwoFA.Status.On {
			// delete QR image from disk
			if lib.FileExist(configSecurity.TwoFA.PathQR + "/" + data2FA.Image) {
				err := os.Remove(configSecurity.TwoFA.PathQR + "/" + data2FA.Image)
				if err != nil {
					log.WithError(err).Error("error code: 1052.42")
				}
			}

			// delete secrets from memory
			service.DelMem2FA(claims.AuthID)

			// DB: 2FA ON, abort setup
			httpResponse.Message = "twoFA: " + configSecurity.TwoFA.Status.On
			httpStatusCode = http.StatusBadRequest
			return
		}
	}

	// step 5: encrypt (AES-256) secret using hash of user's pass
	keyMainCipherByte, err := lib.Encrypt(otpByte, data2FA.PassSHA)
	if err != nil {
		log.WithError(err).Error("error code: 1052.51")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 6: generate recovery key
	keyRecovery := uuid.NewString()
	keyRecovery = strings.ReplaceAll(keyRecovery, "-", "")
	keyRecovery = keyRecovery[len(keyRecovery)-configSecurity.TwoFA.Digits:]
	keyRecoveryHash, err := service.GetHash([]byte(keyRecovery))
	if err != nil {
		log.WithError(err).Error("error code: 1052.61")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 7: encrypt secret using hash of recovery key
	keyBackupCipherByte, err := lib.Encrypt(otpByte, keyRecoveryHash)
	if err != nil {
		log.WithError(err).Error("error code: 1052.71")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 8: generate new UUID code
	uuidPlaintext := uuid.NewString()
	uuidPlaintextByte := []byte(uuidPlaintext)
	uuidSHA, err := service.GetHash(uuidPlaintextByte)
	if err != nil {
		log.WithError(err).Error("error code: 1052.81")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	uuidSHAStr := base64.StdEncoding.EncodeToString(uuidSHA)

	uuidEncByte, err := lib.Encrypt(uuidPlaintextByte, keyRecoveryHash)
	if err != nil {
		log.WithError(err).Error("error code: 1052.82")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 9: encode in base64
	twoFA.KeyMain = base64.StdEncoding.EncodeToString(keyMainCipherByte)
	twoFA.KeyBackup = base64.StdEncoding.EncodeToString(keyBackupCipherByte)
	twoFA.UUIDEnc = base64.StdEncoding.EncodeToString(uuidEncByte)

	// step 10: save in DB
	twoFA.UUIDSHA = uuidSHAStr
	twoFA.Status = configSecurity.TwoFA.Status.On
	twoFA.IDAuth = claims.AuthID

	tx := db.Begin()
	txOK := true

	if available {
		twoFA.UpdatedAt = time.Now()

		if err := tx.Save(&twoFA).Error; err != nil {
			tx.Rollback()
			txOK = false
			log.WithError(err).Error("error code: 1052.100")
		} else {
			tx.Commit()
		}
	}

	if !available {
		if err := tx.Create(&twoFA).Error; err != nil {
			tx.Rollback()
			txOK = false
			log.WithError(err).Error("error code: 1052.101")
		} else {
			tx.Commit()
		}
	}

	if !txOK {
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 11: delete QR image
	if lib.FileExist(configSecurity.TwoFA.PathQR + "/" + data2FA.Image) {
		err = os.Remove(configSecurity.TwoFA.PathQR + "/" + data2FA.Image)
		if err != nil {
			log.WithError(err).Error("error code: 1052.111")
		}
	}

	// step 12: delete secrets from memory
	service.DelMem2FA(claims.AuthID)

	// step 13: issue new tokens
	//
	// set 2FA claim
	claims.TwoFA = configSecurity.TwoFA.Status.Verified
	//
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1052.131")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1052.132")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT
	jwtPayload.TwoAuth = claims.TwoFA
	jwtPayload.RecoveryKey = keyRecovery

	httpResponse.Message = jwtPayload
	httpStatusCode = http.StatusOK
	return
}

// Validate2FA handles jobs for controller.Validate2FA
func Validate2FA(claims middleware.MyCustomClaims, authPayload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// check auth validity
	ok := service.ValidateAuthID(claims.AuthID)
	if !ok {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// check preconditions
	//
	// already verified!
	configSecurity := config.GetConfig().Security
	if claims.TwoFA == configSecurity.TwoFA.Status.Verified {
		httpResponse.Message = "twoFA: " + configSecurity.TwoFA.Status.Verified
		httpStatusCode = http.StatusOK
		return
	}
	// user needs to log in again / 2FA is disabled for this account
	if claims.TwoFA != configSecurity.TwoFA.Status.On {
		httpResponse.Message = "unexpected request (1): 2-fa is OFF / log in again"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// start 2FA validation procedure
	//
	// step 1: check if client secret is available in memory
	data2FA, ok := model.InMemorySecret2FA[claims.AuthID]
	if !ok {
		httpResponse.Message = "log in again"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// step 2: check otp length
	authPayload.OTP = lib.RemoveAllSpace(authPayload.OTP)
	if len(authPayload.OTP) != configSecurity.TwoFA.Digits {
		httpResponse.Message = "wrong one-time password"
		httpStatusCode = http.StatusBadRequest
		return
	}

	newProcess := true
	var encryptedMessage []byte

	// step 3: check if revalidation process
	if len(data2FA.Secret) > 0 {
		encryptedMessage = data2FA.Secret
		newProcess = false
	}

	// check DB
	db := database.GetDB()
	twoFA := model.TwoFA{}
	// no record in DB!
	if err := db.Where("id_auth = ?", claims.AuthID).First(&twoFA).Error; err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1053.31")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// 2FA never configured before for this account
		httpResponse.Message = "unexpected request (2): 2-fa is OFF / log in again"
		httpStatusCode = http.StatusBadRequest
		return
	}
	// if 2FA is not ON
	if twoFA.Status != configSecurity.TwoFA.Status.On {
		// 2FA is disabled for this account
		httpResponse.Message = "unexpected request (3): 2-fa is OFF / log in again"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// retrieve encrypted secrets from DB for new validation process
	if newProcess {
		// decode base64 encoded secret key
		keyMainCipherByte, err := base64.StdEncoding.DecodeString(twoFA.KeyMain)
		if err != nil {
			log.WithError(err).Error("error code: 1053.32")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// decrypt (AES-256) secret using hash of user's pass
		keyMainPlaintextByte, err := lib.Decrypt(keyMainCipherByte, data2FA.PassSHA)
		if err != nil {
			log.WithError(err).Error("error code: 1053.33")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		encryptedMessage = keyMainPlaintextByte
	}

	// step 4: validate user-provided OTP
	otpByte, status, err := service.Validate2FA(
		encryptedMessage,
		configSecurity.TwoFA.Issuer,
		authPayload.OTP,
	)
	if err != nil {
		// client provided invalid OTP
		if status == configSecurity.TwoFA.Status.Invalid {
			// save the new secret in memory for future validation procedure
			data2FA.Secret = otpByte
			model.InMemorySecret2FA[claims.AuthID] = data2FA

			// save in DB to protect from accidental data loss
			//
			// encrypt (AES-256) secret using hash of user's pass
			keyMainCipherByte, err := lib.Encrypt(otpByte, data2FA.PassSHA)
			if err != nil {
				log.WithError(err).Error("error code: 1053.41")
			}
			// encode in base64
			twoFA.KeyMain = base64.StdEncoding.EncodeToString(keyMainCipherByte)
			// updateAt
			twoFA.UpdatedAt = time.Now()
			// write in DB
			tx := db.Begin()
			if err := tx.Save(&twoFA).Error; err != nil {
				tx.Rollback()
				log.WithError(err).Error("error code: 1053.42")
			} else {
				tx.Commit()
			}

			// response to the client
			httpResponse.Message = "wrong one-time password"
			httpStatusCode = http.StatusBadRequest
			return
		}

		// internal error
		log.WithError(err).Error("error code: 1053.43")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 5: 2FA validated
	//
	// encrypt (AES-256) secret using hash of user's pass
	keyMainCipherByte, err := lib.Encrypt(otpByte, data2FA.PassSHA)
	if err != nil {
		log.WithError(err).Error("error code: 1053.51")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	// encode in base64
	twoFA.KeyMain = base64.StdEncoding.EncodeToString(keyMainCipherByte)
	// updateAt
	twoFA.UpdatedAt = time.Now()
	// write in DB
	tx := db.Begin()
	if err := tx.Save(&twoFA).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1053.52")
	} else {
		tx.Commit()
	}
	// delete secrets from memory
	service.DelMem2FA(claims.AuthID)
	//
	// set 2FA claim
	claims.TwoFA = configSecurity.TwoFA.Status.Verified
	//
	// issue new tokens
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1053.53")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1053.54")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT
	jwtPayload.TwoAuth = claims.TwoFA

	httpResponse.Message = jwtPayload
	httpStatusCode = http.StatusOK
	return
}

// Deactivate2FA handles jobs for controller.Deactivate2FA
func Deactivate2FA(claims middleware.MyCustomClaims, authPayload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// app security settings
	configSecurity := config.GetConfig().Security

	// token confirms that 2FA is disabled
	if claims.TwoFA == "" || claims.TwoFA == configSecurity.TwoFA.Status.Off {
		httpResponse.Message = "twoFA: " + configSecurity.TwoFA.Status.Off
		httpStatusCode = http.StatusOK
		return
	}

	// find user
	db := database.GetDB()
	v := model.Auth{}
	if err := db.Where("auth_id = ?", claims.AuthID).First(&v).Error; err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1054.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		httpResponse.Message = "user not found"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// verify password
	verifyPass, err := argon2.ComparePasswordAndHash(authPayload.Password, configSecurity.HashSec, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1054.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !verifyPass {
		httpResponse.Message = "wrong credentials"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// get 2FA info from database
	twoFA := model.TwoFA{}

	err = db.Where("id_auth = ?", v.AuthID).First(&twoFA).Error
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1054.3")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// no record in DB
		// set 2FA claim
		claims.TwoFA = ""
	}

	// record found in DB
	if err == nil {
		// 2FA already disabled in DB
		if twoFA.Status == "" || twoFA.Status == configSecurity.TwoFA.Status.Off {
			// set 2FA claim
			claims.TwoFA = twoFA.Status
		}

		// 2FA is active
		if twoFA.Status == configSecurity.TwoFA.Status.On {
			// remove 2FA keys from DB
			twoFA.UpdatedAt = time.Now()
			twoFA.KeyMain = ""
			twoFA.KeyBackup = ""
			twoFA.UUIDSHA = ""
			twoFA.UUIDEnc = ""
			twoFA.Status = configSecurity.TwoFA.Status.Off

			tx := db.Begin()
			if err := tx.Save(&twoFA).Error; err != nil {
				tx.Rollback()
				log.WithError(err).Error("error code: 1054.4")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}
			tx.Commit()

			// set 2FA claim
			claims.TwoFA = twoFA.Status
		}
	}

	// delete 2FA backup codes from DB if exists
	twoFABackup := []model.TwoFABackup{}
	err = db.Where("id_auth = ?", v.AuthID).Find(&twoFABackup).Error
	if err != nil {
		log.WithError(err).Error("error code: 1054.5")
	}
	if err == nil {
		if len(twoFABackup) > 0 {
			tx := db.Begin()
			if err := tx.Delete(&twoFABackup).Error; err != nil {
				tx.Rollback()
				log.WithError(err).Error("error code: 1054.6")
			} else {
				tx.Commit()
			}
		}
	}

	// generate new tokens
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1054.7")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1054.8")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT
	jwtPayload.TwoAuth = claims.TwoFA

	httpResponse.Message = jwtPayload
	httpStatusCode = http.StatusOK
	return
}

// CreateBackup2FA receives task from controller.CreateBackup2FA.
// If 2FA is already enabled for the user, it generates secret
// backup codes for the user.
//
// Required: valid JWT with parameter "twoFA": "verified"
func CreateBackup2FA(claims middleware.MyCustomClaims, authPayload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// check auth validity
	ok := service.ValidateAuthID(claims.AuthID)
	if !ok {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// is 2FA enabled
	configSecurity := config.GetConfig().Security
	if claims.TwoFA != configSecurity.TwoFA.Status.Verified {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// retrieve user auth
	db := database.GetDB()
	v := model.Auth{}
	if err := db.Where("auth_id = ?", claims.AuthID).First(&v).Error; err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1055.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		httpResponse.Message = "user not found"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// step 1: verify user pass
	verifyPass, err := argon2.ComparePasswordAndHash(authPayload.Password, configSecurity.HashSec, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1055.11")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !verifyPass {
		httpResponse.Message = "wrong credentials"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// step 2: generate 10 backup codes
	var codes [10]string
	for i := 0; i < len(codes); i++ {
		for {
			code, err := service.GenerateCode(16)
			if err != nil {
				continue // retry generating a valid code
			}
			codes[i] = code
			break // exit the inner loop when a code is generated successfully
		}
	}

	// step 3: generate hashes of the codes
	var codeHashes [len(codes)]string
	for i := 0; i < len(codes); i++ {
		codeHash, err := service.CalcHash(
			[]byte(codes[i]),
			configSecurity.Blake2bSec,
		)
		if err != nil {
			log.WithError(err).Error("error code: 1055.31")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		codeHashes[i] = hex.EncodeToString(codeHash)
	}

	// step 4: delete all existing codes of this user
	twoFABackup := []model.TwoFABackup{}

	if err := db.Where("id_auth = ?", claims.AuthID).Find(&twoFABackup).Error; err != nil {
		log.WithError(err).Error("error code: 1055.41")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if len(twoFABackup) > 0 {
		tx := db.Begin()
		if err := tx.Delete(&twoFABackup).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1055.42")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
		tx.Commit()
	}

	// step 5: save the hashes in database
	twoFABackup = []model.TwoFABackup{} // reset
	timeNow := time.Now()

	for i := 0; i < len(codeHashes); i++ {
		backup := model.TwoFABackup{}
		backup.CreatedAt = timeNow
		backup.CodeHash = codeHashes[i]
		backup.IDAuth = claims.AuthID
		twoFABackup = append(twoFABackup, backup)
	}

	tx := db.Begin()
	if err := tx.Create(&twoFABackup).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1055.51")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	// return the plaintext codes to the user
	httpResponse.Message = codes
	httpStatusCode = http.StatusOK
	return
}

// ValidateBackup2FA receives task from controller.ValidateBackup2FA.
// User with 2FA enabled account can verify using their secret backup
// code when they do not have access to their OTP generator app or
// device.
//
// Required: valid JWT with parameter "twoFA": "on"
func ValidateBackup2FA(claims middleware.MyCustomClaims, authPayload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// check auth validity
	ok := service.ValidateAuthID(claims.AuthID)
	if !ok {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// check preconditions
	//
	// already verified!
	configSecurity := config.GetConfig().Security
	if claims.TwoFA == configSecurity.TwoFA.Status.Verified {
		// JWT: 2FA verified
		httpResponse.Message = "twoFA: " + configSecurity.TwoFA.Status.Verified
		httpStatusCode = http.StatusOK
		return
	}
	// user needs to log in again / 2FA is disabled for this account
	if claims.TwoFA != configSecurity.TwoFA.Status.On {
		httpResponse.Message = "unexpected request (1): 2-fa is OFF / log in again"
		httpStatusCode = http.StatusBadRequest
		return
	}

	authPayload.OTP = strings.TrimSpace(authPayload.OTP)
	if authPayload.OTP == "" {
		httpResponse.Message = "required 2-fa backup code"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// retrieve existing 2FA backup codes
	db := database.GetDB()
	twoFABackup := []model.TwoFABackup{}

	if err := db.Where("id_auth = ?", claims.AuthID).Find(&twoFABackup).Error; err != nil {
		log.WithError(err).Error("error code: 1056.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if len(twoFABackup) == 0 {
		httpResponse.Message = "user has no unused valid backup code"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// calculate hash of the given code
	otpByte, err := service.CalcHash(
		[]byte(authPayload.OTP),
		configSecurity.Blake2bSec,
	)
	if err != nil {
		log.WithError(err).Error("error code: 1056.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	otpHash := hex.EncodeToString(otpByte)

	// compare with existing valid hashes
	isOtpValid := false
	validOtpID := uint64(0)
	for _, backup := range twoFABackup {
		if backup.CodeHash == otpHash {
			isOtpValid = true
			validOtpID = backup.ID
			break
		}
	}

	if !isOtpValid {
		httpResponse.Message = "invalid 2-fa backup code"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// delete used code from database
	twoFABackupUsed := model.TwoFABackup{}
	twoFABackupUsed.ID = validOtpID
	tx := db.Begin()
	if err := tx.Delete(&twoFABackupUsed).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1056.3")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	// set 2FA claim
	claims.TwoFA = configSecurity.TwoFA.Status.Verified
	//
	// issue new tokens
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1056.4")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1056.5")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT
	jwtPayload.TwoAuth = claims.TwoFA

	httpResponse.Message = jwtPayload
	httpStatusCode = http.StatusOK
	return
}
