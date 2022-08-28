package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"
)

// PasswordForgot sends secret code for resetting a forgotten password
func PasswordForgot(c *gin.Context) {
	payload := struct {
		Email string `json:"email"`
	}{}

	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	// check email format + perform mx lookup
	payload.Email = strings.TrimSpace(payload.Email)
	if !lib.ValidateEmail(payload.Email) {
		renderer.Render(c, gin.H{"msg": "wrong email address"}, http.StatusBadRequest)
		return
	}

	// find user
	v, err := service.GetUserByEmail(payload.Email)
	if err != nil {
		renderer.Render(c, gin.H{"msg": "user not found"}, http.StatusNotFound)
		return
	}

	// is email already verified
	if v.VerifyEmail != model.EmailVerified {
		renderer.Render(c, gin.H{"msg": "email not verified yet"}, http.StatusBadRequest)
		return
	}

	// send email with secret code
	if !sendEmail(v.Email, model.EmailTypePassRecovery) {
		renderer.Render(c, gin.H{"msg": "sending password recovery email not possible"}, http.StatusBadRequest)
		return
	}

	renderer.Render(c, gin.H{"msg": "sent password recovery email"}, http.StatusOK)
}
