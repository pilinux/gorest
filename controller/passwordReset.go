package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"
)

// PasswordForgot sends secret code for resetting a forgotten password
func PasswordForgot(c *gin.Context) {
	email := model.AuthPayload{}

	if err := c.ShouldBindJSON(&email); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.PasswordForgot(email)

	renderer.Render(c, resp, statusCode)
}

// PasswordRecover resets a forgotten password
func PasswordRecover(c *gin.Context) {
	payload := model.AuthPayload{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.PasswordRecover(payload)

	renderer.Render(c, resp, statusCode)
}

// PasswordUpdate - change password in logged-in state
func PasswordUpdate(c *gin.Context) {
	// get claims
	claims := service.GetClaims(c)

	payload := model.AuthPayload{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.PasswordUpdate(claims, payload)

	renderer.Render(c, resp, statusCode)
}
