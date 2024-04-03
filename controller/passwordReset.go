package controller

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"
)

// PasswordForgot sends secret code for resetting a forgotten password
//
// dependency: relational database, Redis, email service, password recovery service
//
// Accepted JSON payload:
//
// `{"email":"..."}`
func PasswordForgot(c *gin.Context) {
	// delete existing auth cookie if present
	_, errAccessJWT := c.Cookie("accessJWT")
	_, errRefreshJWT := c.Cookie("refreshJWT")
	if errAccessJWT == nil || errRefreshJWT == nil {
		configSecurity := config.GetConfig().Security
		c.SetCookie(
			"accessJWT",
			"",
			-1,
			configSecurity.AuthCookiePath,
			configSecurity.AuthCookieDomain,
			configSecurity.AuthCookieSecure,
			configSecurity.AuthCookieHTTPOnly,
		)
		c.SetCookie(
			"refreshJWT",
			"",
			-1,
			configSecurity.AuthCookiePath,
			configSecurity.AuthCookieDomain,
			configSecurity.AuthCookieSecure,
			configSecurity.AuthCookieHTTPOnly,
		)
	}

	// verify that RDBMS is enabled in .env
	if !config.IsRDBMS() {
		renderer.Render(c, gin.H{"message": "relational database not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that Redis is enabled in .env
	if !config.IsRedis() {
		renderer.Render(c, gin.H{"message": "Redis not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that email service is enabled in .env
	if !config.IsEmailService() {
		renderer.Render(c, gin.H{"message": "email service not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that password recovery service is enabled in .env
	if !config.IsPassRecoveryService() {
		renderer.Render(c, gin.H{"message": "password recovery service not enabled"}, http.StatusNotImplemented)
		return
	}

	email := model.AuthPayload{}

	if err := c.ShouldBindJSON(&email); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.PasswordForgot(email)

	renderer.Render(c, resp, statusCode)
}

// PasswordRecover resets a forgotten password
//
// dependency: relational database, Redis
//
// Accepted JSON payload:
//
// `{"secretCode":"...", "passNew":"...", "passRepeat":"...", "recoveryKey":"..."}`
//
// - `recoveryKey` is required if 2FA is enabled for the user account
func PasswordRecover(c *gin.Context) {
	// delete existing auth cookie if present
	_, errAccessJWT := c.Cookie("accessJWT")
	_, errRefreshJWT := c.Cookie("refreshJWT")
	if errAccessJWT == nil || errRefreshJWT == nil {
		configSecurity := config.GetConfig().Security
		c.SetCookie(
			"accessJWT",
			"",
			-1,
			configSecurity.AuthCookiePath,
			configSecurity.AuthCookieDomain,
			configSecurity.AuthCookieSecure,
			configSecurity.AuthCookieHTTPOnly,
		)
		c.SetCookie(
			"refreshJWT",
			"",
			-1,
			configSecurity.AuthCookiePath,
			configSecurity.AuthCookieDomain,
			configSecurity.AuthCookieSecure,
			configSecurity.AuthCookieHTTPOnly,
		)
	}

	// verify that RDBMS is enabled in .env
	if !config.IsRDBMS() {
		renderer.Render(c, gin.H{"message": "relational database not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that Redis is enabled in .env
	if !config.IsRedis() {
		renderer.Render(c, gin.H{"message": "Redis not enabled"}, http.StatusNotImplemented)
		return
	}

	payload := model.AuthPayload{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.PasswordRecover(payload)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

// PasswordUpdate - change password in logged-in state
//
// dependency: relational database, JWT
//
// Accepted JSON payload:
//
// `{"password":"...", "passNew":"...", "passRepeat":"..."}`
func PasswordUpdate(c *gin.Context) {
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

	// get claims
	claims := service.GetClaims(c)

	payload := model.AuthPayload{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.PasswordUpdate(claims, payload)

	renderer.Render(c, resp, statusCode)
}
