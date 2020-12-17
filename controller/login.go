package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piLinux/GoREST/database/model"
	"github.com/piLinux/GoREST/lib/middleware"
	"github.com/piLinux/GoREST/service"
)

// var db = database.GetDB()

// LoginPayload ...
type LoginPayload struct {
	Email    string `json:"Email"`
	Password string `json:"Password"`
}

// Login ...
func Login(ctx *gin.Context) {
	var payload LoginPayload
	if err := ctx.BindJSON(&payload); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	u, err := service.GetUserByEmail(payload.Email)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if u.Password != model.HashPass(payload.Password) {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	jwtValue, err := middleware.GetJWT(u.ID, u.Email)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"jwt": jwtValue})
}
