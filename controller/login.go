package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piLinux/GoREST/database/model"
	"github.com/piLinux/GoREST/lib/middleware"
	"github.com/piLinux/GoREST/service"
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

	if v.Password != model.HashPass(payload.Password) {
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
