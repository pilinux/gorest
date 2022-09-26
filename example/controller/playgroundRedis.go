package controller

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"

	grenderer "github.com/pilinux/gorest/lib/renderer"

	"github.com/pilinux/gorest/example/database/model"
	"github.com/pilinux/gorest/example/handler"
)

// RedisCreate - SET key
func RedisCreate(c *gin.Context) {
	data := model.RedisData{}
	if err := c.ShouldBindJSON(&data); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisCreate(data)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		grenderer.Render(c, resp, statusCode)
		return
	}

	grenderer.Render(c, resp.Message, statusCode)
}

// RedisRead - GET key
func RedisRead(c *gin.Context) {
	data := model.RedisData{}
	if err := c.ShouldBindJSON(&data); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisRead(data)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		grenderer.Render(c, resp, statusCode)
		return
	}

	grenderer.Render(c, resp.Message, statusCode)
}

// RedisDelete - DEL key
func RedisDelete(c *gin.Context) {
	data := model.RedisData{}
	if err := c.ShouldBindJSON(&data); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisDelete(data)

	grenderer.Render(c, resp, statusCode)
}

// RedisCreateHash - SET hashes
func RedisCreateHash(c *gin.Context) {
	data := model.RedisDataHash{}
	if err := c.ShouldBindJSON(&data); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisCreateHash(data)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		grenderer.Render(c, resp, statusCode)
		return
	}

	grenderer.Render(c, resp.Message, statusCode)
}

// RedisReadHash - GET hashes
func RedisReadHash(c *gin.Context) {
	data := model.RedisDataHash{}
	if err := c.ShouldBindJSON(&data); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisReadHash(data)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		grenderer.Render(c, resp, statusCode)
		return
	}

	grenderer.Render(c, resp.Message, statusCode)
}

// RedisDeleteHash - DEL hashes
func RedisDeleteHash(c *gin.Context) {
	data := model.RedisDataHash{}
	if err := c.ShouldBindJSON(&data); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisDeleteHash(data)

	grenderer.Render(c, resp, statusCode)
}
