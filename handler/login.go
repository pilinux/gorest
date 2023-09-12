package handler

import (
	"net/http"
	"strings"

	"github.com/pilinux/argon2"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
	"github.com/pilinux/gorest/service"
)

// Login receives tasks from controller.Login.
// After authentication, it returns new access and refresh tokens.
func Login(payload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	payload.Email = strings.TrimSpace(payload.Email)
	if !lib.ValidateEmail(payload.Email) {
		httpResponse.Message = "wrong email address"
		httpStatusCode = http.StatusBadRequest
		return
	}

	v, err := service.GetUserByEmail(payload.Email, false)
	if err != nil {
		if err.Error() != database.RecordNotFound {
			// db read error
			log.WithError(err).Error("error code: 1013.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

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

	verifyPass, err := argon2.ComparePasswordAndHash(payload.Password, configSecurity.HashSec, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1013.2")
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
	// claims.Email
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
		err := db.Where("id_auth = ?", v.AuthID).First(&twoFA).Error
		if err != nil {
			if err.Error() != database.RecordNotFound {
				// db read error
				log.WithError(err).Error("error code: 1013.3")
				httpResponse.Message = "internal server error"
				httpStatusCode = http.StatusInternalServerError
				return
			}
		}
		if err == nil {
			claims.TwoFA = twoFA.Status

			// 2FA ON
			if twoFA.Status == configSecurity.TwoFA.Status.On {
				// hash user's pass
				hashPass, err := service.GetHash([]byte(payload.Password))
				if err != nil {
					log.WithError(err).Error("error code: 1013.4")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}

				// save the hashed pass in memory for OTP validation step
				data2FA := model.Secret2FA{}
				data2FA.PassSHA = hashPass
				model.InMemorySecret2FA[claims.AuthID] = data2FA
			}
		}
	}

	// issue new tokens
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1013.5")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1013.6")
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

// Refresh receives tasks from controller.Refresh and
// returns new pair of tokens (access and refresh tokens).
func Refresh(claims middleware.MyCustomClaims) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// check validity
	ok := service.ValidateAuthID(claims.AuthID)
	if !ok {
		httpResponse.Message = "access denied"
		httpStatusCode = http.StatusUnauthorized
		return
	}

	// issue new tokens
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1014.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1014.2")
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
