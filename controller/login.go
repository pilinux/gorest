package controller

import (
	"crypto/sha256"
	"net/http"
	"os"
	"strings"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"

	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// LoginPayload ...
type LoginPayload struct {
	Email    string `json:"Email"`
	Password string `json:"Password"`
}

// Login - issue new JWTs after user:pass verification
func Login(c *gin.Context) {
	var payload LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

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
	if config.SecurityConfigAll.Must2FA == config.Activated {
		db := database.GetDB()
		twoFA := model.TwoFA{}

		// have the user configured 2FA
		if err := db.Where("id_auth = ?", v.AuthID).First(&twoFA).Error; err == nil {
			// 2FA ON
			if twoFA.Status == config.SecurityConfigAll.TwoFA.Status.On {
				claims.TwoFA = twoFA.Status
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
	if claims.TwoFA == config.SecurityConfigAll.TwoFA.Status.Verified || claims.TwoFA == config.SecurityConfigAll.TwoFA.Status.On {
		renderer.Render(c, gin.H{"msg": "2-fa activated already"}, http.StatusOK)
		return
	}

	// is 2FA disabled/never configured before
	db := database.GetDB()
	twoFA := model.TwoFA{}
	// err != nil => never configured before
	// err == nil => 2FA disabled
	if err := db.Where("id_auth = ?", claims.AuthID).First(&twoFA).Error; err == nil {
		if twoFA.Status == config.SecurityConfigAll.TwoFA.Status.On {
			// 2FA ON
			renderer.Render(c, gin.H{"msg": "2-fa activated already"}, http.StatusBadRequest)
			return
		}
	}

	// start OTP setup procedure
	//
	// step 1: verify user pass
	pass := struct {
		Password string `json:"Password"`
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
		config.SecurityConfigAll.TwoFA.Issuer,
		config.SecurityConfigAll.TwoFA.Crypto,
		config.SecurityConfigAll.TwoFA.Digits,
	)
	if err != nil {
		log.WithError(err).Error("error code: 1032")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	// step 3: encode QR in bytes
	qrByte, err := lib.NewQR(
		otpByte,
		config.SecurityConfigAll.TwoFA.Issuer,
	)
	if err != nil {
		log.WithError(err).Error("error code: 1033")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	// step 4: generate QR in PNG format and save on disk
	img, err := lib.ByteToPNG(qrByte, config.SecurityConfigAll.TwoFA.PathQR)
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
		// delete old QR image
		err := os.Remove(config.SecurityConfigAll.TwoFA.PathQR + data2FA.Image)
		if err != nil {
			log.WithError(err).Error("error code: 1035")
		}
	}

	// step 7: save the secrets in RAM for validation
	data2FA.PassSHA = hashPass[:]
	data2FA.Secret = otpByte
	data2FA.Image = img
	model.InMemorySecret2FA[claims.AuthID] = data2FA

	// serve the QR to the client
	c.File(config.SecurityConfigAll.TwoFA.PathQR + img)
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

// delMem2FA - delete secrets from memory
func delMem2FA(authID uint64) {
	delete(model.InMemorySecret2FA, authID)
}
