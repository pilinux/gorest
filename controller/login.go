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

// Login - issue new JWTs after user:pass verification
func Login(c *gin.Context) {
	var payload model.AuthPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Login(payload)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

// Refresh - issue new JWTs after validation
func Refresh(c *gin.Context) {
	// get claims
	claims := service.GetClaims(c)

	resp, statusCode := handler.Refresh(claims)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}
