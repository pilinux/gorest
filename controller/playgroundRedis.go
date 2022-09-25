package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"
)

// RedisCreate - SET key
func RedisCreate(c *gin.Context) {
	data := model.RedisData{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisCreate(data)

	renderer.Render(c, resp, statusCode)
}

// RedisRead - GET key
func RedisRead(c *gin.Context) {
	data := model.RedisData{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisRead(data)

	renderer.Render(c, resp, statusCode)
}

// RedisDelete - DEL key
func RedisDelete(c *gin.Context) {
	data := model.RedisData{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisDelete(data)

	renderer.Render(c, resp, statusCode)
}

// RedisCreateHash - SET hashes
func RedisCreateHash(c *gin.Context) {
	data := model.RedisDataHash{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisCreateHash(data)

	renderer.Render(c, resp, statusCode)
}

// RedisReadHash - GET hashes
func RedisReadHash(c *gin.Context) {
	data := model.RedisDataHash{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisReadHash(data)

	renderer.Render(c, resp, statusCode)
}

// RedisDeleteHash - DEL hashes
func RedisDeleteHash(c *gin.Context) {
	data := model.RedisDataHash{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisReadHash(data)

	renderer.Render(c, resp, statusCode)
}
