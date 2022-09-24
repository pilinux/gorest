package controller

import (
	"net/http"
	"strings"

	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"

	"github.com/gin-gonic/gin"
)

// GetUsers - GET /users
func GetUsers(c *gin.Context) {
	resp, statusCode := handler.GetUsers()

	renderer.Render(c, resp, statusCode)
}

// GetUser - GET /users/:id
func GetUser(c *gin.Context) {
	id := strings.TrimSpace(c.Params.ByName("id"))

	resp, statusCode := handler.GetUser(id)

	renderer.Render(c, resp, statusCode)
}

// CreateUser - POST /users
func CreateUser(c *gin.Context) {
	userIDAuth := c.GetUint64("authID")
	user := model.User{}

	// bind JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.CreateUser(userIDAuth, user)

	renderer.Render(c, resp, statusCode)
}

// UpdateUser - PUT /users
func UpdateUser(c *gin.Context) {
	userIDAuth := c.GetUint64("authID")
	user := model.User{}

	// bind JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.UpdateUser(userIDAuth, user)

	renderer.Render(c, resp, statusCode)
}

// AddHobby - PUT /users/hobbies
func AddHobby(c *gin.Context) {
	userIDAuth := c.GetUint64("authID")
	hobby := model.Hobby{}

	// bind JSON
	if err := c.ShouldBindJSON(&hobby); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.AddHobby(userIDAuth, hobby)

	renderer.Render(c, resp, statusCode)
}
