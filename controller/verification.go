package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"
)

// VerifyEmail - verify email address
func VerifyEmail(c *gin.Context) {
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

	resp, statusCode := handler.VerifyEmail(payload)

	renderer.Render(c, resp, statusCode)
}

// CreateVerificationEmail issues new verification code upon request
func CreateVerificationEmail(c *gin.Context) {
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

	resp, statusCode := handler.CreateVerificationEmail(payload)

	renderer.Render(c, resp, statusCode)
}
