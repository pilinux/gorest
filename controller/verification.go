package controller

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"
)

// VerifyEmail - verify email address
// dependency: email verification service, Redis
func VerifyEmail(c *gin.Context) {
	// delete existing auth cookie if present
	_, errAccessJWT := c.Cookie("accessJWT")
	_, errRefreshJWT := c.Cookie("refreshJWT")
	if errAccessJWT == nil || errRefreshJWT == nil {
		c.SetCookie("accessJWT", "", -1, "", "", true, true)
		c.SetCookie("refreshJWT", "", -1, "", "", true, true)
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

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp, statusCode)
}

// CreateVerificationEmail issues new verification code upon request
// dependency: email service, email verification service, Redis
func CreateVerificationEmail(c *gin.Context) {
	// delete existing auth cookie if present
	_, errAccessJWT := c.Cookie("accessJWT")
	_, errRefreshJWT := c.Cookie("refreshJWT")
	if errAccessJWT == nil || errRefreshJWT == nil {
		c.SetCookie("accessJWT", "", -1, "", "", true, true)
		c.SetCookie("refreshJWT", "", -1, "", "", true, true)
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

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp, statusCode)
}
