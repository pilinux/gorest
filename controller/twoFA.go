package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
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

	renderer.Render(c, resp, statusCode)
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

	resp, statusCode := handler.Setup2FA(claims, otp)

	if statusCode >= 400 {
		renderer.Render(c, resp, statusCode)
		return
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

	if statusCode >= 400 {
		renderer.Render(c, resp, statusCode)
		return
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

	resp, statusCode := handler.Validate2FA(claims, password)

	if statusCode >= 400 {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}
