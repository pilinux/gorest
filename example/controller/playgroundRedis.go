package controller

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"

	grenderer "github.com/pilinux/gorest/lib/renderer"

	"github.com/pilinux/gorest/example/database/model"
	"github.com/pilinux/gorest/example/handler"
)

// RedisCreate handles SET key operations.
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

// RedisRead handles GET key operations.
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

// RedisDelete handles DEL key operations.
func RedisDelete(c *gin.Context) {
	data := model.RedisData{}
	if err := c.ShouldBindJSON(&data); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisDelete(data)

	grenderer.Render(c, resp, statusCode)
}

// RedisCreateHash handles SET hash operations.
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

// RedisReadHash handles GET hash operations.
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

// RedisDeleteHash handles DEL hash operations.
func RedisDeleteHash(c *gin.Context) {
	data := model.RedisDataHash{}
	if err := c.ShouldBindJSON(&data); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.RedisDeleteHash(data)

	grenderer.Render(c, resp, statusCode)
}
