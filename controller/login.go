package controller

import (
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	"github.com/pilinux/gorest/lib/middleware"
	"github.com/pilinux/gorest/service"
)

// LoginPayload ...
type LoginPayload struct {
	Email    string `json:"Email"`
	Password string `json:"Password"`
}

// Login ...
func Login(c *gin.Context) {
	var payload LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	v, err := service.GetUserByEmail(payload.Email)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	verifyPass, err := argon2id.ComparePasswordAndHash(payload.Password, v.Password)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if !verifyPass {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	jwtValue, err := middleware.GetJWT(v.AuthID, v.Email)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"JWT": jwtValue})
}
