// Package controller contains all the controllers
// of the application
package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"
)

// CreateUserAuth - POST /register
func CreateUserAuth(c *gin.Context) {
	auth := model.Auth{}

	// bind JSON
	if err := c.ShouldBindJSON(&auth); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.CreateUserAuth(auth)

	renderer.Render(c, resp, statusCode)
}
