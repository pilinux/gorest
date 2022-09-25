// Package controller contains all the controllers
// of the application
package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	grenderer "github.com/pilinux/gorest/lib/renderer"

	"github.com/pilinux/gorest/example/database/model"
	"github.com/pilinux/gorest/example/handler"
)

// GetUsers - GET /users
func GetUsers(c *gin.Context) {
	resp, statusCode := handler.GetUsers()

	grenderer.Render(c, resp, statusCode)
}

// GetUser - GET /users/:id
func GetUser(c *gin.Context) {
	id := strings.TrimSpace(c.Params.ByName("id"))

	resp, statusCode := handler.GetUser(id)

	grenderer.Render(c, resp, statusCode)
}

// CreateUser - POST /users
func CreateUser(c *gin.Context) {
	userIDAuth := c.GetUint64("authID")
	user := model.User{}

	// bind JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		grenderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.CreateUser(userIDAuth, user)

	grenderer.Render(c, resp, statusCode)
}

// UpdateUser - PUT /users
func UpdateUser(c *gin.Context) {
	userIDAuth := c.GetUint64("authID")
	user := model.User{}

	// bind JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		grenderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.UpdateUser(userIDAuth, user)

	grenderer.Render(c, resp, statusCode)
}

// AddHobby - PUT /users/hobbies
func AddHobby(c *gin.Context) {
	userIDAuth := c.GetUint64("authID")
	hobby := model.Hobby{}

	// bind JSON
	if err := c.ShouldBindJSON(&hobby); err != nil {
		grenderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.AddHobby(userIDAuth, hobby)

	grenderer.Render(c, resp, statusCode)
}
