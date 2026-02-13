package handler

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/pilinux/argon2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
	"github.com/pilinux/gorest/service"
)

// Login authenticates a user and returns a new access/refresh token pair.
//
// If email verification is enabled, it requires the account email to be verified.
// If 2FA is enabled and configured for the user, it prepares in-memory state for
// the OTP validation step.
func Login(payload model.AuthPayload) (httpResponse model.HTTPResponse, httpStatusCode int) {
	payload.Email = strings.TrimSpace(payload.Email)
	if !lib.ValidateEmail(payload.Email) {
		httpResponse.Message = "wrong email address"
		httpStatusCode = http.StatusBadRequest
		return
	}

	v, err := service.GetUserByEmail(payload.Email, false)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
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
			if !errors.Is(err, gorm.ErrRecordNotFound) {
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
				var key []byte

				// check if KeySalt is available
				if twoFA.KeySalt != "" {
					salt, err := base64.StdEncoding.DecodeString(twoFA.KeySalt)
					if err != nil {
						log.WithError(err).Error("error code: 1013.4.1")
						httpResponse.Message = "internal server error"
						httpStatusCode = http.StatusInternalServerError
						return
					}
					key = lib.GetArgon2Key([]byte(payload.Password), salt, 32)
				} else {
					// if KeySalt is not available, it means the user configured 2FA before
					// the KeySalt feature was introduced in v1.10.5.
					// in v1.11.0, support for users without KeySalt is dropped,
					// and there is no fallback to old 2FA mechanism.
					// please check release notes of v1.10.5 to solve this issue for
					// existing users before migrating to v1.11.x.
					log.WithFields(
						log.Fields{
							"authID": v.AuthID,
							"reason": "missing KeySalt for 2FA",
						}).
						Error("error code: 1013.4.2")
					httpResponse.Message = "internal server error"
					httpStatusCode = http.StatusInternalServerError
					return
				}

				// save the hashed pass (key) in memory for OTP validation step
				data2FA := model.Secret2FA{}
				data2FA.PassHash = key
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

// Refresh issues a new access/refresh token pair for an authenticated session.
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
