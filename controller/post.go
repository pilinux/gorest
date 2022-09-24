package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"

	"github.com/gin-gonic/gin"
)

// GetPosts - GET /posts
func GetPosts(c *gin.Context) {
	resp, statusCode := handler.GetPosts()

	renderer.Render(c, resp, statusCode)
}

// GetPost - GET /posts/:id
func GetPost(c *gin.Context) {
	id := strings.TrimSpace(c.Params.ByName("id"))

	resp, statusCode := handler.GetPost(id)

	if statusCode >= 400 {
		errorMsg := model.ErrorMsg{}
		errorMsg.HTTPCode = statusCode
		errorMsg.Message = fmt.Sprintf("%v", resp.Result)

		renderer.Render(c, errorMsg, statusCode, "error.html")
		return
	}

	renderer.Render(c, resp, statusCode, "read-article.html")
}

// CreatePost - POST /posts
func CreatePost(c *gin.Context) {
	userIDAuth := c.GetUint64("authID")
	post := model.Post{}

	// bind JSON
	if err := c.ShouldBindJSON(&post); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.CreatePost(userIDAuth, post)

	renderer.Render(c, resp, statusCode)
}

// UpdatePost - PUT /posts/:id
func UpdatePost(c *gin.Context) {
	userIDAuth := c.GetUint64("authID")
	id := strings.TrimSpace(c.Params.ByName("id"))
	post := model.Post{}

	// bind JSON
	if err := c.ShouldBindJSON(&post); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.UpdatePost(userIDAuth, id, post)

	renderer.Render(c, resp, statusCode)
}

// DeletePost - DELETE /posts/:id
func DeletePost(c *gin.Context) {
	userIDAuth := c.GetUint64("authID")
	id := strings.TrimSpace(c.Params.ByName("id"))

	resp, statusCode := handler.DeletePost(userIDAuth, id)

	renderer.Render(c, resp, statusCode)
}
