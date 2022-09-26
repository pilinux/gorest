package controller

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"

	grenderer "github.com/pilinux/gorest/lib/renderer"

	"github.com/pilinux/gorest/example/database/model"
	"github.com/pilinux/gorest/example/handler"
)

// GetPosts - GET /posts
func GetPosts(c *gin.Context) {
	resp, statusCode := handler.GetPosts()

	grenderer.Render(c, resp, statusCode)
}

// GetPost - GET /posts/:id
func GetPost(c *gin.Context) {
	id := strings.TrimSpace(c.Params.ByName("id"))

	resp, statusCode := handler.GetPost(id)

	if statusCode >= 400 {
		errorMsg := model.ErrorMsg{}
		errorMsg.HTTPCode = statusCode
		errorMsg.Message = fmt.Sprintf("%v", resp.Message)

		grenderer.Render(c, errorMsg, statusCode, "error.html")
		return
	}

	grenderer.Render(c, resp.Message, statusCode, "read-article.html")
}

// CreatePost - POST /posts
func CreatePost(c *gin.Context) {
	userIDAuth := c.GetUint64("authID")
	post := model.Post{}

	// bind JSON
	if err := c.ShouldBindJSON(&post); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.CreatePost(userIDAuth, post)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		grenderer.Render(c, resp, statusCode)
		return
	}

	grenderer.Render(c, resp.Message, statusCode)
}

// UpdatePost - PUT /posts/:id
func UpdatePost(c *gin.Context) {
	userIDAuth := c.GetUint64("authID")
	id := strings.TrimSpace(c.Params.ByName("id"))
	post := model.Post{}

	// bind JSON
	if err := c.ShouldBindJSON(&post); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.UpdatePost(userIDAuth, id, post)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		grenderer.Render(c, resp, statusCode)
		return
	}

	grenderer.Render(c, resp.Message, statusCode)
}

// DeletePost - DELETE /posts/:id
func DeletePost(c *gin.Context) {
	userIDAuth := c.GetUint64("authID")
	id := strings.TrimSpace(c.Params.ByName("id"))

	resp, statusCode := handler.DeletePost(userIDAuth, id)

	grenderer.Render(c, resp, statusCode)
}
