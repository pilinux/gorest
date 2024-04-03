package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"
)

// VerifyEmail - verify email address of a newly registered user account
//
// dependency: email verification service, Redis
//
// Accepted JSON payload:
//
// `{"verificationCode":"..."}`
func VerifyEmail(c *gin.Context) {
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

	// verify that email verification service is enabled in .env
	if !config.IsEmailVerificationService() {
		renderer.Render(c, gin.H{"message": "email verification service not enabled"}, http.StatusNotImplemented)
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

	resp, statusCode := handler.VerifyEmail(payload)

	renderer.Render(c, resp, statusCode)
}

// CreateVerificationEmail issues new verification code upon request
//
// dependency: email service, email verification service, Redis
//
// Accepted JSON payload:
//
// `{"email":"...", "password":"..."}`
func CreateVerificationEmail(c *gin.Context) {
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

	// verify that email service is enabled in .env
	if !config.IsEmailService() {
		renderer.Render(c, gin.H{"message": "email service not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that email verification service is enabled in .env
	if !config.IsEmailVerificationService() {
		renderer.Render(c, gin.H{"message": "email verification service not enabled"}, http.StatusNotImplemented)
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

	resp, statusCode := handler.CreateVerificationEmail(payload)

	renderer.Render(c, resp, statusCode)
}

// VerifyUpdatedEmail - verify request to modify user's email address
//
// dependency: email verification service, relational database, redis
//
// Accepted JSON payload:
//
// `{"verificationCode":"..."}`
func VerifyUpdatedEmail(c *gin.Context) {
	// verify that email verification service is enabled in .env
	if !config.IsEmailVerificationService() {
		renderer.Render(c, gin.H{"message": "email verification service not enabled"}, http.StatusNotImplemented)
		return
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

	resp, statusCode := handler.VerifyUpdatedEmail(payload)

	renderer.Render(c, resp, statusCode)
}

// GetUnverifiedEmail - if any email is yet to be verified, return it to the logged-in user
//
// When this email is verified, it will replace the existing active email of the user.
//
// dependency: email verification service, relational database, JWT
func GetUnverifiedEmail(c *gin.Context) {
	// verify that email verification service is enabled in .env
	if !config.IsEmailVerificationService() {
		renderer.Render(c, gin.H{"message": "email verification service not enabled"}, http.StatusNotImplemented)
		return
	}

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

	resp, statusCode := handler.GetUnverifiedEmail(claims)

	renderer.Render(c, resp, statusCode)
}

// ResendVerificationCodeToModifyActiveEmail issues new verification code upon request
//
// dependency: email service, email verification service, Redis,
// relational database, JWT
func ResendVerificationCodeToModifyActiveEmail(c *gin.Context) {
	// verify that email service is enabled in .env
	if !config.IsEmailService() {
		renderer.Render(c, gin.H{"message": "email service not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that email verification service is enabled in .env
	if !config.IsEmailVerificationService() {
		renderer.Render(c, gin.H{"message": "email verification service not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that Redis is enabled in .env
	if !config.IsRedis() {
		renderer.Render(c, gin.H{"message": "Redis not enabled"}, http.StatusNotImplemented)
		return
	}

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

	resp, statusCode := handler.ResendVerificationCodeToModifyActiveEmail(claims)

	renderer.Render(c, resp, statusCode)
}
