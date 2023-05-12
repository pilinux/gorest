package controller

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/middleware"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"
)

// Setup2FA - get secret to activate 2FA
// possible for accounts without 2FA-ON
func Setup2FA(c *gin.Context) {
	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	password := model.AuthPayload{}
	if err := c.ShouldBindJSON(&password); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Setup2FA(claims, password)

	if statusCode != 201 {
		renderer.Render(c, resp, statusCode)
		return
	}

	// serve the QR code
	c.File(fmt.Sprintf("%v", resp.Message))
}

// Activate2FA - activate 2FA upon validation
// possible for accounts without 2FA-ON
func Activate2FA(c *gin.Context) {
	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	otp := model.AuthPayload{}
	if err := c.ShouldBindJSON(&otp); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Activate2FA(claims, otp)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	// set cookie if the feature is enabled in app settings
	configSecurity := config.GetConfig().Security
	if configSecurity.AuthCookieActivate {
		tokens, ok := resp.Message.(middleware.JWTPayload)
		if ok {
			c.SetSameSite(configSecurity.AuthCookieSameSite)
			c.SetCookie(
				"accessJWT",
				tokens.AccessJWT,
				middleware.JWTParams.AccessKeyTTL*60,
				configSecurity.AuthCookiePath,
				configSecurity.AuthCookieDomain,
				configSecurity.AuthCookieSecure,
				configSecurity.AuthCookieHTTPOnly,
			)
			c.SetCookie(
				"refreshJWT",
				tokens.RefreshJWT,
				middleware.JWTParams.RefreshKeyTTL*60,
				configSecurity.AuthCookiePath,
				configSecurity.AuthCookieDomain,
				configSecurity.AuthCookieSecure,
				configSecurity.AuthCookieHTTPOnly,
			)

			if !configSecurity.ServeJwtAsResBody {
				tokens.AccessJWT = ""
				tokens.RefreshJWT = ""
				resp.Message = tokens
			}
		}

		if !ok {
			log.Error("error code: 1041.1")
			resp.Message = "failed to prepare auth cookie"
			statusCode = http.StatusInternalServerError
		}
	}

	renderer.Render(c, resp.Message, statusCode)
}

// Validate2FA - issue new JWTs upon 2FA validation
// required for accounts with 2FA-ON
func Validate2FA(c *gin.Context) {
	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	otp := model.AuthPayload{}
	if err := c.ShouldBindJSON(&otp); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Validate2FA(claims, otp)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	// set cookie if the feature is enabled in app settings
	configSecurity := config.GetConfig().Security
	if configSecurity.AuthCookieActivate {
		tokens, ok := resp.Message.(middleware.JWTPayload)
		if ok {
			c.SetSameSite(configSecurity.AuthCookieSameSite)
			c.SetCookie(
				"accessJWT",
				tokens.AccessJWT,
				middleware.JWTParams.AccessKeyTTL*60,
				configSecurity.AuthCookiePath,
				configSecurity.AuthCookieDomain,
				configSecurity.AuthCookieSecure,
				configSecurity.AuthCookieHTTPOnly,
			)
			c.SetCookie(
				"refreshJWT",
				tokens.RefreshJWT,
				middleware.JWTParams.RefreshKeyTTL*60,
				configSecurity.AuthCookiePath,
				configSecurity.AuthCookieDomain,
				configSecurity.AuthCookieSecure,
				configSecurity.AuthCookieHTTPOnly,
			)

			if !configSecurity.ServeJwtAsResBody {
				resp.Message = "2-fa verified"
			}
		}

		if !ok {
			log.Error("error code: 1051.1")
			resp.Message = "failed to prepare auth cookie"
			statusCode = http.StatusInternalServerError
		}
	}

	renderer.Render(c, resp.Message, statusCode)
}

// Deactivate2FA - disable 2FA for user account
func Deactivate2FA(c *gin.Context) {
	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	password := model.AuthPayload{}
	if err := c.ShouldBindJSON(&password); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Deactivate2FA(claims, password)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	// set cookie if the feature is enabled in app settings
	configSecurity := config.GetConfig().Security
	if configSecurity.AuthCookieActivate {
		tokens, ok := resp.Message.(middleware.JWTPayload)
		if ok {
			c.SetSameSite(configSecurity.AuthCookieSameSite)
			c.SetCookie(
				"accessJWT",
				tokens.AccessJWT,
				middleware.JWTParams.AccessKeyTTL*60,
				configSecurity.AuthCookiePath,
				configSecurity.AuthCookieDomain,
				configSecurity.AuthCookieSecure,
				configSecurity.AuthCookieHTTPOnly,
			)
			c.SetCookie(
				"refreshJWT",
				tokens.RefreshJWT,
				middleware.JWTParams.RefreshKeyTTL*60,
				configSecurity.AuthCookiePath,
				configSecurity.AuthCookieDomain,
				configSecurity.AuthCookieSecure,
				configSecurity.AuthCookieHTTPOnly,
			)

			if !configSecurity.ServeJwtAsResBody {
				tokens.AccessJWT = ""
				tokens.RefreshJWT = ""
				resp.Message = tokens
			}
		}

		if !ok {
			log.Error("error code: 1036.1")
			resp.Message = "failed to prepare auth cookie"
			statusCode = http.StatusInternalServerError
		}
	}

	renderer.Render(c, resp.Message, statusCode)
}
