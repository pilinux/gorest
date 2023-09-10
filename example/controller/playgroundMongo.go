package controller

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"

	grenderer "github.com/pilinux/gorest/lib/renderer"

	"github.com/pilinux/gorest/example/database/model"
	"github.com/pilinux/gorest/example/handler"
)

// MongoCreateOne - create one document
func MongoCreateOne(c *gin.Context) {
	data := model.Geocoding{}
	if err := c.ShouldBindJSON(&data); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.MongoCreateOne(data)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		grenderer.Render(c, resp, statusCode)
		return
	}

	grenderer.Render(c, resp.Message, statusCode)
}

// MongoGetAll - get all documents
func MongoGetAll(c *gin.Context) {
	resp, statusCode := handler.MongoGetAll()

	grenderer.Render(c, resp, statusCode)
}

// MongoGetByID - find one document by ID
func MongoGetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Params.ByName("id"))

	resp, statusCode := handler.MongoGetByID(id)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		grenderer.Render(c, resp, statusCode)
		return
	}

	grenderer.Render(c, resp.Message, statusCode)
}

// MongoGetByFilter - find documents using filter
func MongoGetByFilter(c *gin.Context) {
	req := model.Geocoding{}
	if err := c.ShouldBindJSON(&req); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.MongoGetByFilter(req)

	grenderer.Render(c, resp, statusCode)
}

// MongoUpdateByID - update a document
//
// - edit existing fields
// - add new fields
// - do not remove any existing field
func MongoUpdateByID(c *gin.Context) {
	req := model.Geocoding{}
	if err := c.ShouldBindJSON(&req); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.MongoUpdateByID(req)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		grenderer.Render(c, resp, statusCode)
		return
	}

	grenderer.Render(c, resp.Message, statusCode)
}

// MongoDeleteFieldByID - delete existing field(s) from a document
func MongoDeleteFieldByID(c *gin.Context) {
	req := model.Geocoding{}
	if err := c.ShouldBindJSON(&req); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.MongoDeleteFieldByID(req)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		grenderer.Render(c, resp, statusCode)
		return
	}

	grenderer.Render(c, resp.Message, statusCode)
}

// MongoDeleteByID - delete one document by ID
func MongoDeleteByID(c *gin.Context) {
	id := strings.TrimSpace(c.Params.ByName("id"))

	resp, statusCode := handler.MongoDeleteByID(id)

	grenderer.Render(c, resp, statusCode)
}
