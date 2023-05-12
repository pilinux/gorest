package controller

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"
)

// PasswordForgot sends secret code for resetting a forgotten password
func PasswordForgot(c *gin.Context) {
	// delete existing auth cookie if present
	_, errAccessJWT := c.Cookie("accessJWT")
	_, errRefreshJWT := c.Cookie("refreshJWT")
	if errAccessJWT == nil || errRefreshJWT == nil {
		c.SetCookie("accessJWT", "", -1, "", "", true, true)
		c.SetCookie("refreshJWT", "", -1, "", "", true, true)
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
func PasswordRecover(c *gin.Context) {
	// delete existing auth cookie if present
	_, errAccessJWT := c.Cookie("accessJWT")
	_, errRefreshJWT := c.Cookie("refreshJWT")
	if errAccessJWT == nil || errRefreshJWT == nil {
		c.SetCookie("accessJWT", "", -1, "", "", true, true)
		c.SetCookie("refreshJWT", "", -1, "", "", true, true)
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
func PasswordUpdate(c *gin.Context) {
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
