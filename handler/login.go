package handler

import (
	"crypto/sha256"
	"net/http"
	"strings"

	"github.com/alexedwards/argon2id"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
	"github.com/pilinux/gorest/service"
)

// Login handles jobs for controller.Login
func Login(payload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	payload.Email = strings.TrimSpace(payload.Email)
	if !lib.ValidateEmail(payload.Email) {
		httpResponse.Message = "wrong email address"
		httpStatusCode = http.StatusBadRequest
		return
	}

	v, err := service.GetUserByEmail(payload.Email)
	if err != nil {
		httpResponse.Message = "email not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	// app settings
	configSecurity := config.GetConfig().Security

	// check whether email verification is required
	if configSecurity.VerifyEmail {
		if v.VerifyEmail != model.EmailVerified {
			httpResponse.Message = "email verification required"
			httpStatusCode = http.StatusUnauthorized
			return
		}
	}

	verifyPass, err := argon2id.ComparePasswordAndHash(payload.Password, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1011")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if !verifyPass {
		httpResponse.Message = "wrong credentials"
		httpStatusCode = http.StatusUnauthorized
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
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1013")
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

// Refresh handles jobs for controller.Refresh
func Refresh(claims middleware.MyCustomClaims) (httpResponse model.HTTPResponse, httpStatusCode int) {

	// check validity
	ok := service.ValidateUserID(claims.AuthID, claims.Email)
	if !ok {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// issue new tokens
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1021")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1022")
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
