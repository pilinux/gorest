package controller

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"

	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// LoginPayload ...
type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login - issue new JWTs after user:pass verification
func Login(c *gin.Context) {
	var payload LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	payload.Email = strings.TrimSpace(payload.Email)
	if !lib.ValidateEmail(payload.Email) {
		renderer.Render(c, gin.H{"msg": "wrong email address"}, http.StatusBadRequest)
		return
	}

	v, err := service.GetUserByEmail(payload.Email)
	if err != nil {
		renderer.Render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	verifyPass, err := argon2id.ComparePasswordAndHash(payload.Password, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1011")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	if !verifyPass {
		renderer.Render(c, gin.H{"msg": "wrong credentials"}, http.StatusUnauthorized)
		return
	}

	// custom claims
	claims := middleware.MyCustomClaims{}
	claims.AuthID = v.AuthID
	claims.Email = v.Email
	// claims.Role
	// claims.Scope
	// claims.TwoFA
	// claims.SiteLan
	// claims.Custom1
	// claims.Custom2

	// when 2FA is enabled for this application (ACTIVATE_2FA=yes)
	configSecurity := config.GetConfig().Security
	if configSecurity.Must2FA == config.Activated {
		db := database.GetDB()
		twoFA := model.TwoFA{}

		// have the user configured 2FA
		if err := db.Where("id_auth = ?", v.AuthID).First(&twoFA).Error; err == nil {
			// 2FA ON
			if twoFA.Status == configSecurity.TwoFA.Status.On {
				claims.TwoFA = twoFA.Status

				// hash user's pass in sha256
				hashPass := sha256.Sum256([]byte(payload.Password))

				// save the hashed pass in memory for OTP validation step
				data2FA := model.Secret2FA{}
				data2FA.PassSHA = hashPass[:]
				model.InMemorySecret2FA[claims.AuthID] = data2FA
			}
		}
	}

	// issue new tokens
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1012")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1013")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT
	renderer.Render(c, jwtPayload, http.StatusOK)
}

// Refresh - issue new JWTs after validation
func Refresh(c *gin.Context) {
	// get claims
	claims := getClaims(c)

	// check validity
	ok := validateUserID(claims.AuthID, claims.Email)
	if !ok {
		renderer.Render(c, gin.H{"msg": "access denied"}, http.StatusUnauthorized)
		return
	}

	// issue new tokens
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1021")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1022")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT
	renderer.Render(c, jwtPayload, http.StatusOK)
}

// Setup2FA - get secret to activate 2FA
// possible for accounts without 2FA-ON
func Setup2FA(c *gin.Context) {
	// get claims
	claims := getClaims(c)

	// check user validity
	ok := validateUserID(claims.AuthID, claims.Email)
	if !ok {
		renderer.Render(c, gin.H{"msg": "access denied"}, http.StatusUnauthorized)
		return
	}

	// 2FA already enabled
	configSecurity := config.GetConfig().Security
	if claims.TwoFA == configSecurity.TwoFA.Status.Verified || claims.TwoFA == configSecurity.TwoFA.Status.On {
		renderer.Render(c, gin.H{"msg": "2-fa activated already"}, http.StatusOK)
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
			renderer.Render(c, gin.H{"msg": "2-fa activated already"}, http.StatusBadRequest)
			return
		}
	}

	// start OTP setup procedure
	//
	// step 1: verify user pass
	pass := struct {
		Password string `json:"password"`
	}{}
	// bind JSON
	if err := c.ShouldBindJSON(&pass); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}
	v, err := service.GetUserByEmail(claims.Email)
	if err != nil {
		renderer.Render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}
	verifyPass, err := argon2id.ComparePasswordAndHash(pass.Password, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1031")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	if !verifyPass {
		renderer.Render(c, gin.H{"msg": "wrong credentials"}, http.StatusUnauthorized)
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
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	// step 3: encode QR in bytes
	qrByte, err := lib.NewQR(
		otpByte,
		configSecurity.TwoFA.Issuer,
	)
	if err != nil {
		log.WithError(err).Error("error code: 1033")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	// step 4: generate QR in PNG format and save on disk
	img, err := lib.ByteToPNG(qrByte, configSecurity.TwoFA.PathQR)
	if err != nil {
		log.WithError(err).Error("error code: 1034")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	// step 5: hash user's pass in sha256
	hashPass := sha256.Sum256([]byte(pass.Password))

	// step 6: check if client secret is available in memory
	data2FA, ok := model.InMemorySecret2FA[claims.AuthID]
	if ok {
		// delete old QR image if available
		if lib.FileExist(configSecurity.TwoFA.PathQR + data2FA.Image) {
			err := os.Remove(configSecurity.TwoFA.PathQR + data2FA.Image)
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
	c.File(configSecurity.TwoFA.PathQR + img)
}

// Activate2FA - activate 2FA upon validation
// possible for accounts without 2FA-ON
func Activate2FA(c *gin.Context) {
	// get claims
	claims := getClaims(c)

	// check user validity
	ok := validateUserID(claims.AuthID, claims.Email)
	if !ok {
		renderer.Render(c, gin.H{"msg": "access denied"}, http.StatusUnauthorized)
		return
	}

	configSecurity := config.GetConfig().Security
	if claims.TwoFA == configSecurity.TwoFA.Status.Verified || claims.TwoFA == configSecurity.TwoFA.Status.On {
		renderer.Render(c, gin.H{"msg": "2-fa activated already"}, http.StatusBadRequest)
		return
	}

	// start 2FA activation procedure
	//
	// step 1: check if client secret is available in memory
	data2FA, ok := model.InMemorySecret2FA[claims.AuthID]
	if !ok {
		renderer.Render(c, gin.H{"msg": "request for a new 2-fa secret"}, http.StatusBadRequest)
		return
	}

	// step 2: bind JSON
	userInput := struct {
		Input string `json:"otp"`
	}{}
	if err := c.ShouldBindJSON(&userInput); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}
	userInput.Input = lib.RemoveAllSpace(userInput.Input)
	if len(userInput.Input) != configSecurity.TwoFA.Digits {
		renderer.Render(c, gin.H{"msg": "wrong one-time password"}, http.StatusUnauthorized)
		return
	}

	// step 3: validate user-provided OTP
	otpByte, status, err := validate2FA(
		data2FA.Secret,
		configSecurity.TwoFA.Issuer,
		userInput.Input,
	)
	if err != nil {
		// client provided invalid OTP
		if status == configSecurity.TwoFA.Status.Invalid {
			// save the secret with failed attempt in memory for future validation procedure
			data2FA.Secret = otpByte
			model.InMemorySecret2FA[claims.AuthID] = data2FA

			renderer.Render(c, gin.H{"msg": "wrong OTP"}, http.StatusUnauthorized)
			return
		}

		// internal error
		log.WithError(err).Error("error code: 1041")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
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
			if lib.FileExist(configSecurity.TwoFA.PathQR + data2FA.Image) {
				err := os.Remove(configSecurity.TwoFA.PathQR + data2FA.Image)
				if err != nil {
					log.WithError(err).Error("error code: 1042")
				}
			}

			// delete secrets from memory
			delMem2FA(claims.AuthID)

			renderer.Render(c, gin.H{"msg": "2-fa activated already"}, http.StatusBadRequest)
			return
		}
	}

	// step 5: encrypt (AES-256) secret using hash of user's pass
	keyMainInByte, err := lib.Encrypt(otpByte, data2FA.PassSHA)
	if err != nil {
		log.WithError(err).Error("error code: 1043")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	// step 6: generate recovery key
	keyRecovery := uuid.NewString()
	keyRecovery = keyRecovery[len(keyRecovery)-6:]
	keyRecoveryHash := sha256.Sum256([]byte(keyRecovery))

	// step 7: encrypt secret using hash of recovery key
	keyMBackupInByte, err := lib.Encrypt(otpByte, keyRecoveryHash[:])
	if err != nil {
		log.WithError(err).Error("error code: 1044")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	// step 8: encode in base64
	twoFA.KeyMain = base64.StdEncoding.EncodeToString(keyMainInByte)
	twoFA.KeyBackup = base64.StdEncoding.EncodeToString(keyMBackupInByte)

	// step 9: save in DB
	twoFA.Status = configSecurity.TwoFA.Status.On
	twoFA.IDAuth = claims.AuthID

	tx := db.Begin()
	txOK := true

	if available {
		twoFA.UpdatedAt = time.Now().Local()

		if err := tx.Save(&twoFA).Error; err != nil {
			tx.Rollback()
			txOK = false
			log.WithError(err).Error("error code: 1045")
		} else {
			tx.Commit()
		}
	}

	if !available {
		if err := tx.Create(&twoFA).Error; err != nil {
			tx.Rollback()
			txOK = false
			log.WithError(err).Error("error code: 1046")
		} else {
			tx.Commit()
		}
	}

	if !txOK {
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	// step 10: delete QR image
	if lib.FileExist(configSecurity.TwoFA.PathQR + data2FA.Image) {
		err = os.Remove(configSecurity.TwoFA.PathQR + data2FA.Image)
		if err != nil {
			log.WithError(err).Error("error code: 1047")
		}
	}

	// step 11: delete secrets from memory
	delMem2FA(claims.AuthID)

	// step 12: issue new tokens
	//
	// set 2FA claim
	claims.TwoFA = configSecurity.TwoFA.Status.Verified
	//
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1048")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1049")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	// send response to the client
	response := struct {
		AccessJWT   string `json:"accessJWT"`
		RefreshJWT  string `json:"refreshJWT"`
		TwoAuth     string `json:"twoFA"`
		RecoveryKey string `json:"recoveryKey"`
	}{}

	response.AccessJWT = accessJWT
	response.RefreshJWT = refreshJWT
	response.TwoAuth = configSecurity.TwoFA.Status.On
	response.RecoveryKey = keyRecovery

	renderer.Render(c, response, http.StatusOK)
}

// Validate2FA - issue new JWTs upon 2FA validation
// required for accounts with 2FA-ON
func Validate2FA(c *gin.Context) {
	// get claims
	claims := getClaims(c)

	// check user validity
	ok := validateUserID(claims.AuthID, claims.Email)
	if !ok {
		renderer.Render(c, gin.H{"msg": "access denied"}, http.StatusUnauthorized)
		return
	}

	// check preconditions
	//
	// already verified!
	configSecurity := config.GetConfig().Security
	if claims.TwoFA == configSecurity.TwoFA.Status.Verified {
		renderer.Render(
			c,
			gin.H{"msg": configSecurity.TwoFA.Status.Verified},
			http.StatusOK,
		)
		return
	}
	// user needs to log in again / 2FA is disabled for this account
	if claims.TwoFA != configSecurity.TwoFA.Status.On {
		renderer.Render(
			c,
			gin.H{"msg": "unexpected request (1): 2-fa is OFF / log in again"},
			http.StatusBadRequest,
		)
		return
	}

	// start 2FA validation procedure
	//
	// step 1: check if client secret is available in memory
	data2FA, ok := model.InMemorySecret2FA[claims.AuthID]
	if !ok {
		renderer.Render(c, gin.H{"msg": "log in again"}, http.StatusUnauthorized)
		return
	}

	// step 2: bind JSON
	userInput := struct {
		Input string `json:"otp"`
	}{}
	if err := c.ShouldBindJSON(&userInput); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}
	userInput.Input = lib.RemoveAllSpace(userInput.Input)
	if len(userInput.Input) != configSecurity.TwoFA.Digits {
		renderer.Render(c, gin.H{"msg": "wrong one-time password"}, http.StatusUnauthorized)
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
		renderer.Render(
			c,
			gin.H{"msg": "unexpected request (2): 2-fa is OFF / log in again"},
			http.StatusBadRequest,
		)
		return
	}
	// if 2FA is not ON
	if twoFA.Status != configSecurity.TwoFA.Status.On {
		renderer.Render(
			c,
			gin.H{"msg": "unexpected request (3): 2-fa is OFF / log in again"},
			http.StatusBadRequest,
		)
		return
	}

	// retrieve encrypted secrets from DB for new validation process
	if newProcess {
		// decode base64 encoded secret key
		cipherInByte, err := base64.StdEncoding.DecodeString(twoFA.KeyMain)
		if err != nil {
			log.WithError(err).Error("error code: 1051")
			renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
			return
		}

		// decrypt (AES-256) secret using hash of user's pass
		secretInByte, err := lib.Decrypt(cipherInByte, data2FA.PassSHA)
		if err != nil {
			log.WithError(err).Error("error code: 1052")
			renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
			return
		}

		encryptedMessage = secretInByte
	}

	// step 4: validate user-provided OTP
	otpByte, status, err := validate2FA(
		encryptedMessage,
		configSecurity.TwoFA.Issuer,
		userInput.Input,
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
			keyMainInByte, err := lib.Encrypt(otpByte, data2FA.PassSHA)
			if err != nil {
				log.WithError(err).Error("error code: 1053")
			}
			// encode in base64
			twoFA.KeyMain = base64.StdEncoding.EncodeToString(keyMainInByte)
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
			renderer.Render(c, gin.H{"msg": "wrong OTP"}, http.StatusUnauthorized)
			return
		}

		// internal error
		log.WithError(err).Error("error code: 1055")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	// step 5: 2FA validated
	//
	// encrypt (AES-256) secret using hash of user's pass
	keyMainInByte, err := lib.Encrypt(otpByte, data2FA.PassSHA)
	if err != nil {
		log.WithError(err).Error("error code: 1056")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	// encode in base64
	twoFA.KeyMain = base64.StdEncoding.EncodeToString(keyMainInByte)
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
	delMem2FA(claims.AuthID)
	//
	// set 2FA claim
	claims.TwoFA = configSecurity.TwoFA.Status.Verified
	//
	// issue new tokens
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1058")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1059")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT
	renderer.Render(c, jwtPayload, http.StatusOK)
}

// getClaims - get JWT custom claims
func getClaims(c *gin.Context) middleware.MyCustomClaims {
	// get claims
	claims := middleware.MyCustomClaims{
		AuthID:  c.GetUint64("authID"),
		Email:   c.GetString("email"),
		Role:    c.GetString("role"),
		Scope:   c.GetString("scope"),
		TwoFA:   c.GetString("tfa"),
		SiteLan: c.GetString("siteLan"),
		Custom1: c.GetString("custom1"),
		Custom2: c.GetString("custom2"),
	}

	return claims
}

// validateUserID - check whether authID or email is missing
func validateUserID(authID uint64, email string) bool {
	email = strings.TrimSpace(email)
	return authID != 0 && email != ""
}

// validate2FA validates user-provided OTP
func validate2FA(encryptedMessage []byte, issuer string, userInput string) ([]byte, string, error) {
	configSecurity := config.GetConfig().Security
	otpByte, err := lib.ValidateTOTP(encryptedMessage, issuer, userInput)
	// client provided invalid OTP / internal error
	if err != nil {
		// client provided invalid OTP
		if len(otpByte) > 0 {
			return otpByte, configSecurity.TwoFA.Status.Invalid, err
		}

		// internal error
		return []byte{}, "", err
	}

	// validated
	return otpByte, configSecurity.TwoFA.Status.Verified, nil
}

// delMem2FA - delete secrets from memory
func delMem2FA(authID uint64) {
	delete(model.InMemorySecret2FA, authID)
}
