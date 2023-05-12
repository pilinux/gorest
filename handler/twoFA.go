package handler

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
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
	// check user validity
	ok := service.ValidateUserID(claims.AuthID, claims.Email)
	if !ok {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// 2FA already enabled
	configSecurity := config.GetConfig().Security
	if claims.TwoFA == configSecurity.TwoFA.Status.Verified || claims.TwoFA == configSecurity.TwoFA.Status.On {
		httpResponse.Message = "2-fa activated already"
		httpStatusCode = http.StatusOK
		return
	}

	// is 2FA disabled/never configured before
	db := database.GetDB()
	twoFA := model.TwoFA{}
	// err != nil => never configured before
	// err == nil => 2FA disabled
	if err := db.Where("id_auth = ?", claims.AuthID).First(&twoFA).Error; err == nil {
		if twoFA.Status == configSecurity.TwoFA.Status.On {
			// 2FA ON
			httpResponse.Message = "2-fa activated already, log in again"
			httpStatusCode = http.StatusBadRequest
			return
		}
	}

	// start OTP setup procedure
	//
	// step 1: verify user pass
	v, err := service.GetUserByEmail(claims.Email)
	if err != nil {
		httpResponse.Message = "user not found"
		httpStatusCode = http.StatusUnauthorized
		return
	}
	verifyPass, err := argon2id.ComparePasswordAndHash(authPayload.Password, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1031")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !verifyPass {
		httpResponse.Message = "wrong credentials"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// step 2: create new TOTP object
	otpByte, err := lib.NewTOTP(
		claims.Email,
		configSecurity.TwoFA.Issuer,
		configSecurity.TwoFA.Crypto,
		configSecurity.TwoFA.Digits,
	)
	if err != nil {
		log.WithError(err).Error("error code: 1032")
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
		log.WithError(err).Error("error code: 1033")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 4: generate QR in PNG format and save on disk
	img, err := lib.ByteToPNG(qrByte, configSecurity.TwoFA.PathQR+"/")
	if err != nil {
		log.WithError(err).Error("error code: 1034")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 5: hash user's pass in sha256
	hashPass := sha256.Sum256([]byte(authPayload.Password))

	// step 6: check if client secret is available in memory
	data2FA, ok := model.InMemorySecret2FA[claims.AuthID]
	if ok {
		// delete old QR image if available
		if lib.FileExist(configSecurity.TwoFA.PathQR + "/" + data2FA.Image) {
			err := os.Remove(configSecurity.TwoFA.PathQR + "/" + data2FA.Image)
			if err != nil {
				log.WithError(err).Error("error code: 1035")
			}
		}
	}

	// step 7: save the secrets in memory for validation
	data2FA.PassSHA = hashPass[:]
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
	// check user validity
	ok := service.ValidateUserID(claims.AuthID, claims.Email)
	if !ok {
		httpResponse.Message = "validation failed - access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	configSecurity := config.GetConfig().Security
	if claims.TwoFA == configSecurity.TwoFA.Status.On {
		httpResponse.Message = "2-fa activated already, log in again"
		httpStatusCode = http.StatusBadRequest
		return
	}
	if claims.TwoFA == configSecurity.TwoFA.Status.Verified {
		httpResponse.Message = "2-fa activated already"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// start 2FA activation procedure
	//
	// step 1: check if client secret is available in memory
	data2FA, ok := model.InMemorySecret2FA[claims.AuthID]
	if !ok {
		httpResponse.Message = "request for a new 2-fa secret"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// step 2: check otp length
	authPayload.OTP = lib.RemoveAllSpace(authPayload.OTP)
	if len(authPayload.OTP) != configSecurity.TwoFA.Digits {
		httpResponse.Message = "wrong one-time password"
		httpStatusCode = http.StatusUnauthorized
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
			httpStatusCode = http.StatusUnauthorized
			return
		}

		// internal error
		log.WithError(err).Error("error code: 1041")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 4: check DB
	db := database.GetDB()
	twoFA := model.TwoFA{}
	available := false
	if err := db.Where("id_auth = ?", claims.AuthID).First(&twoFA).Error; err == nil {
		// record found in DB
		available = true

		// 2FA already activated!
		if twoFA.Status == configSecurity.TwoFA.Status.On {
			// delete QR image from disk
			if lib.FileExist(configSecurity.TwoFA.PathQR + "/" + data2FA.Image) {
				err := os.Remove(configSecurity.TwoFA.PathQR + "/" + data2FA.Image)
				if err != nil {
					log.WithError(err).Error("error code: 1042")
				}
			}

			// delete secrets from memory
			service.DelMem2FA(claims.AuthID)

			httpResponse.Message = "2-fa activated already, log in again"
			httpStatusCode = http.StatusBadRequest
			return
		}
	}

	// step 5: encrypt (AES-256) secret using hash of user's pass
	keyMainCipherByte, err := lib.Encrypt(otpByte, data2FA.PassSHA)
	if err != nil {
		log.WithError(err).Error("error code: 1043")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 6: generate recovery key
	keyRecovery := uuid.NewString()
	keyRecovery = strings.ReplaceAll(keyRecovery, "-", "")
	keyRecovery = keyRecovery[len(keyRecovery)-configSecurity.TwoFA.Digits:]
	keyRecoveryHash := sha256.Sum256([]byte(keyRecovery))

	// step 7: encrypt secret using hash of recovery key
	keyBackupCipherByte, err := lib.Encrypt(otpByte, keyRecoveryHash[:])
	if err != nil {
		log.WithError(err).Error("error code: 1044")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 8: generate new UUID code
	uuidPlaintext := uuid.NewString()
	uuidPlaintextByte := []byte(uuidPlaintext)
	uuidSHA256 := sha256.Sum256(uuidPlaintextByte)
	uuidSHA := base64.StdEncoding.EncodeToString(uuidSHA256[:])

	uuidEncByte, err := lib.Encrypt(uuidPlaintextByte, keyRecoveryHash[:])
	if err != nil {
		log.WithError(err).Error("error code: 1045")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 9: encode in base64
	twoFA.KeyMain = base64.StdEncoding.EncodeToString(keyMainCipherByte)
	twoFA.KeyBackup = base64.StdEncoding.EncodeToString(keyBackupCipherByte)
	twoFA.UUIDEnc = base64.StdEncoding.EncodeToString(uuidEncByte)

	// step 10: save in DB
	twoFA.UUIDSHA = uuidSHA
	twoFA.Status = configSecurity.TwoFA.Status.On
	twoFA.IDAuth = claims.AuthID

	tx := db.Begin()
	txOK := true

	if available {
		twoFA.UpdatedAt = time.Now().Local()

		if err := tx.Save(&twoFA).Error; err != nil {
			tx.Rollback()
			txOK = false
			log.WithError(err).Error("error code: 1046")
		} else {
			tx.Commit()
		}
	}

	if !available {
		if err := tx.Create(&twoFA).Error; err != nil {
			tx.Rollback()
			txOK = false
			log.WithError(err).Error("error code: 1047")
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
			log.WithError(err).Error("error code: 1048")
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
		log.WithError(err).Error("error code: 1049")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1050")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT
	jwtPayload.TwoAuth = configSecurity.TwoFA.Status.Verified
	jwtPayload.RecoveryKey = keyRecovery

	httpResponse.Message = jwtPayload
	httpStatusCode = http.StatusOK
	return
}

// Validate2FA handles jobs for controller.Validate2FA
func Validate2FA(claims middleware.MyCustomClaims, authPayload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// check user validity
	ok := service.ValidateUserID(claims.AuthID, claims.Email)
	if !ok {
		httpResponse.Message = "validation failed - access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// check preconditions
	//
	// already verified!
	configSecurity := config.GetConfig().Security
	if claims.TwoFA == configSecurity.TwoFA.Status.Verified {
		httpResponse.Message = configSecurity.TwoFA.Status.Verified
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
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// step 2: check otp length
	authPayload.OTP = lib.RemoveAllSpace(authPayload.OTP)
	if len(authPayload.OTP) != configSecurity.TwoFA.Digits {
		httpResponse.Message = "wrong one-time password"
		httpStatusCode = http.StatusUnauthorized
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
		httpResponse.Message = "unexpected request (2): 2-fa is OFF / log in again"
		httpStatusCode = http.StatusBadRequest
		return
	}
	// if 2FA is not ON
	if twoFA.Status != configSecurity.TwoFA.Status.On {
		httpResponse.Message = "unexpected request (3): 2-fa is OFF / log in again"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// retrieve encrypted secrets from DB for new validation process
	if newProcess {
		// decode base64 encoded secret key
		keyMainCipherByte, err := base64.StdEncoding.DecodeString(twoFA.KeyMain)
		if err != nil {
			log.WithError(err).Error("error code: 1051")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// decrypt (AES-256) secret using hash of user's pass
		keyMainPlaintextByte, err := lib.Decrypt(keyMainCipherByte, data2FA.PassSHA)
		if err != nil {
			log.WithError(err).Error("error code: 1052")
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
				log.WithError(err).Error("error code: 1053")
			}
			// encode in base64
			twoFA.KeyMain = base64.StdEncoding.EncodeToString(keyMainCipherByte)
			// updateAt
			twoFA.UpdatedAt = time.Now().Local()
			// write in DB
			tx := db.Begin()
			if err := tx.Save(&twoFA).Error; err != nil {
				tx.Rollback()
				log.WithError(err).Error("error code: 1054")
			} else {
				tx.Commit()
			}

			// response to the client
			httpResponse.Message = "wrong one-time password"
			httpStatusCode = http.StatusUnauthorized
			return
		}

		// internal error
		log.WithError(err).Error("error code: 1055")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// step 5: 2FA validated
	//
	// encrypt (AES-256) secret using hash of user's pass
	keyMainCipherByte, err := lib.Encrypt(otpByte, data2FA.PassSHA)
	if err != nil {
		log.WithError(err).Error("error code: 1056")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	// encode in base64
	twoFA.KeyMain = base64.StdEncoding.EncodeToString(keyMainCipherByte)
	// updateAt
	twoFA.UpdatedAt = time.Now().Local()
	// write in DB
	tx := db.Begin()
	if err := tx.Save(&twoFA).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1057")
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
		log.WithError(err).Error("error code: 1058")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1059")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT

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
		httpResponse.Message = "twoFA is " + configSecurity.TwoFA.Status.Off
		httpStatusCode = http.StatusOK
		return
	}

	// find user
	v, err := service.GetUserByEmail(claims.Email)
	if err != nil {
		httpResponse.Message = "unknown user"
		httpStatusCode = http.StatusNotFound
		return
	}
	// verify password
	verifyPass, err := argon2id.ComparePasswordAndHash(authPayload.Password, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1036")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !verifyPass {
		httpResponse.Message = "wrong credentials"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// get 2FA info from database
	db := database.GetDB()
	twoFA := model.TwoFA{}

	err = db.Where("id_auth = ?", v.AuthID).First(&twoFA).Error
	if err != nil {
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
			twoFA.UpdatedAt = time.Now().Local()
			twoFA.KeyMain = ""
			twoFA.KeyBackup = ""
			twoFA.UUIDSHA = ""
			twoFA.UUIDEnc = ""
			twoFA.Status = configSecurity.TwoFA.Status.Off

			tx := db.Begin()
			if err := tx.Save(&twoFA).Error; err != nil {
				tx.Rollback()
				log.WithError(err).Error("error code: 1037")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}
			tx.Commit()

			// set 2FA claim
			claims.TwoFA = twoFA.Status
		}
	}

	// generate new tokens
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1038")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1039")
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
