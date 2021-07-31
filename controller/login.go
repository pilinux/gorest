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
		render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}
	v, err := service.GetUserByEmail(payload.Email)
	if err != nil {
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	verifyPass, err := argon2id.ComparePasswordAndHash(payload.Password, v.Password)
	if err != nil {
		render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	if !verifyPass {
		render(c, gin.H{"msg": "wrong credentials"}, http.StatusUnauthorized)
		return
	}

	jwtValue, err := middleware.GetJWT(v.AuthID, v.Email)
	if err != nil {
		render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	render(c, gin.H{"JWT": jwtValue}, http.StatusOK)
}
