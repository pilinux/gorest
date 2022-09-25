package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"
)

// MongoCreateOne - create one document
func MongoCreateOne(c *gin.Context) {
	data := model.Geocoding{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.MongoCreateOne(data)

	renderer.Render(c, resp, statusCode)
}

// MongoGetAll - get all documents
func MongoGetAll(c *gin.Context) {
	resp, statusCode := handler.MongoGetAll()

	renderer.Render(c, resp, statusCode)
}

// MongoGetByID - find one document by ID
func MongoGetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Params.ByName("id"))

	resp, statusCode := handler.MongoGetByID(id)

	renderer.Render(c, resp, statusCode)
}

// MongoGetByFilter - find documents using filter
func MongoGetByFilter(c *gin.Context) {
	req := model.Geocoding{}
	if err := c.ShouldBindJSON(&req); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.MongoGetByFilter(req)

	renderer.Render(c, resp, statusCode)
}

// MongoUpdateByID - update a document
// edit existing fields
// add new fields
// do not remove any existing field
func MongoUpdateByID(c *gin.Context) {
	req := model.Geocoding{}
	if err := c.ShouldBindJSON(&req); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.MongoUpdateByID(req)

	renderer.Render(c, resp, statusCode)
}

// MongoDeleteFieldByID - delete existing field(s) from a document
func MongoDeleteFieldByID(c *gin.Context) {
	req := model.Geocoding{}
	if err := c.ShouldBindJSON(&req); err != nil {
		renderer.Render(c, gin.H{"result": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.MongoDeleteFieldByID(req)

	renderer.Render(c, resp, statusCode)
}

// MongoDeleteByID - delete one document by ID
func MongoDeleteByID(c *gin.Context) {
	id := strings.TrimSpace(c.Params.ByName("id"))

	resp, statusCode := handler.MongoDeleteByID(id)

	renderer.Render(c, resp, statusCode)
}
