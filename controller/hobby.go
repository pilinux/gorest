package controller

import (
	"net/http"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib/renderer"

	"github.com/gin-gonic/gin"
)

// GetHobbies - GET /hobbies
func GetHobbies(c *gin.Context) {
	db := database.GetDB()
	hobbies := []model.Hobby{}

	if err := db.Find(&hobbies).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
	} else {
		renderer.Render(c, hobbies, http.StatusOK)
	}
}
