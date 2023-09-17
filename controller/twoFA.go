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
//
// - possible for accounts without 2FA-ON
//
// dependency: relational database, JWT, 2FA service
//
// Accepted JSON payload:
//
// `{"password":"..."}`
func Setup2FA(c *gin.Context) {
	// verify that RDBMS is enabled in .env
	if !config.IsRDBMS() {
		renderer.Render(c, gin.H{"message": "relational database not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that JWT service is enabled in .env
	if !config.IsJWT() {
		renderer.Render(c, gin.H{"message": "JWT service not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that two-factor authentication service is enabled in .env
	if !config.Is2FA() {
		renderer.Render(c, gin.H{"message": "two-factor authentication service not enabled"}, http.StatusNotImplemented)
		return
	}

	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	password := model.AuthPayload{}
	if err := c.ShouldBindJSON(&password); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Setup2FA(claims, password)

	if statusCode != http.StatusCreated {
		renderer.Render(c, resp, statusCode)
		return
	}

	// serve the QR code
	c.File(fmt.Sprintf("%v", resp.Message))
}

// Activate2FA - activate 2FA upon validation
//
// - possible for accounts without 2FA-ON
//
// dependency: relational database, JWT, 2FA service
//
// Accepted JSON payload:
//
// `{"otp":"..."}`
func Activate2FA(c *gin.Context) {
	// verify that RDBMS is enabled in .env
	if !config.IsRDBMS() {
		renderer.Render(c, gin.H{"message": "relational database not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that JWT service is enabled in .env
	if !config.IsJWT() {
		renderer.Render(c, gin.H{"message": "JWT service not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that two-factor authentication service is enabled in .env
	if !config.Is2FA() {
		renderer.Render(c, gin.H{"message": "two-factor authentication service not enabled"}, http.StatusNotImplemented)
		return
	}

	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	otp := model.AuthPayload{}
	if err := c.ShouldBindJSON(&otp); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Activate2FA(claims, otp)

	// 2FA activation failed
	if statusCode != http.StatusOK {
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

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

// Validate2FA - issue new JWTs upon 2FA validation
//
// - required for accounts with 2FA-ON
//
// dependency: relational database, JWT, 2FA service
//
// Accepted JSON payload:
//
// `{"otp":"..."}`
func Validate2FA(c *gin.Context) {
	// verify that RDBMS is enabled in .env
	if !config.IsRDBMS() {
		renderer.Render(c, gin.H{"message": "relational database not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that JWT service is enabled in .env
	if !config.IsJWT() {
		renderer.Render(c, gin.H{"message": "JWT service not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that two-factor authentication service is enabled in .env
	if !config.Is2FA() {
		renderer.Render(c, gin.H{"message": "two-factor authentication service not enabled"}, http.StatusNotImplemented)
		return
	}

	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	otp := model.AuthPayload{}
	if err := c.ShouldBindJSON(&otp); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Validate2FA(claims, otp)

	// 2FA validation failed
	if statusCode != http.StatusOK {
		renderer.Render(c, resp, statusCode)
		return
	}

	// JWT already verified, no need to issue new tokens
	if resp.Message == "twoFA: "+config.GetConfig().Security.TwoFA.Status.Verified {
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
			log.Error("error code: 1042.1")
			resp.Message = "failed to prepare auth cookie"
			statusCode = http.StatusInternalServerError
		}
	}

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

// Deactivate2FA - disable 2FA for user account
//
// dependency: relational database, JWT, 2FA service
//
// Accepted JSON payload:
//
// `{"password":"..."}`
func Deactivate2FA(c *gin.Context) {
	// verify that RDBMS is enabled in .env
	if !config.IsRDBMS() {
		renderer.Render(c, gin.H{"message": "relational database not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that JWT service is enabled in .env
	if !config.IsJWT() {
		renderer.Render(c, gin.H{"message": "JWT service not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that two-factor authentication service is enabled in .env
	if !config.Is2FA() {
		renderer.Render(c, gin.H{"message": "two-factor authentication service not enabled"}, http.StatusNotImplemented)
		return
	}

	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	password := model.AuthPayload{}
	if err := c.ShouldBindJSON(&password); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Deactivate2FA(claims, password)

	// 2FA deactivation failed
	if statusCode != http.StatusOK {
		renderer.Render(c, resp, statusCode)
		return
	}

	// based on JWT, 2FA is already disabled
	if resp.Message == "twoFA: "+config.GetConfig().Security.TwoFA.Status.Off {
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
			log.Error("error code: 1043.1")
			resp.Message = "failed to prepare auth cookie"
			statusCode = http.StatusInternalServerError
		}
	}

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

// CreateBackup2FA - get new set of 2FA backup codes
//
// - 2FA must already be enabled for the user account
//
// dependency: relational database, JWT, 2FA service
//
// Accepted JSON payload:
//
// `{"password":"..."}`
func CreateBackup2FA(c *gin.Context) {
	// verify that RDBMS is enabled in .env
	if !config.IsRDBMS() {
		renderer.Render(c, gin.H{"message": "relational database not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that JWT service is enabled in .env
	if !config.IsJWT() {
		renderer.Render(c, gin.H{"message": "JWT service not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that two-factor authentication service is enabled in .env
	if !config.Is2FA() {
		renderer.Render(c, gin.H{"message": "two-factor authentication service not enabled"}, http.StatusNotImplemented)
		return
	}

	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	password := model.AuthPayload{}
	if err := c.ShouldBindJSON(&password); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.CreateBackup2FA(claims, password)

	renderer.Render(c, resp, statusCode)
}

// ValidateBackup2FA - issue new JWTs upon 2FA validation with backup code
//
// dependency: relational database, JWT, 2FA service
//
// Accepted JSON payload:
//
// `{"otp":"..."}`
func ValidateBackup2FA(c *gin.Context) {
	// verify that RDBMS is enabled in .env
	if !config.IsRDBMS() {
		renderer.Render(c, gin.H{"message": "relational database not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that JWT service is enabled in .env
	if !config.IsJWT() {
		renderer.Render(c, gin.H{"message": "JWT service not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that two-factor authentication service is enabled in .env
	if !config.Is2FA() {
		renderer.Render(c, gin.H{"message": "two-factor authentication service not enabled"}, http.StatusNotImplemented)
		return
	}

	// get claims
	claims := service.GetClaims(c)

	// bind JSON
	authPayload := model.AuthPayload{}
	if err := c.ShouldBindJSON(&authPayload); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.ValidateBackup2FA(claims, authPayload)

	// 2FA validation with backup code failed
	if statusCode != http.StatusOK {
		renderer.Render(c, resp, statusCode)
		return
	}

	// JWT already verified, no need to issue new tokens
	if resp.Message == "twoFA: "+config.GetConfig().Security.TwoFA.Status.Verified {
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
			log.Error("error code: 1044.1")
			resp.Message = "failed to prepare auth cookie"
			statusCode = http.StatusInternalServerError
		}
	}

	renderer.Render(c, resp, statusCode)
}
