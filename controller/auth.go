package controller

import (
	"fmt"
	"net/http"

	"github.com/piLinux/GoREST/database"
	"github.com/piLinux/GoREST/database/model"

	"github.com/gin-gonic/gin"
)

// CreateUserAuth - POST /register
func CreateUserAuth(c *gin.Context) {
	db := database.GetDB()
	auth := model.Auth{}
	createAuth := 0 // default value

	c.ShouldBindJSON(&auth)

	if err := db.Where("email = ?", auth.Email).First(&auth).Error; err == nil {
		createAuth = 1 // email is already registered
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// one unique email for each account
	if createAuth == 0 {
		tx := db.Begin()
		if err := tx.Create(&auth).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
		} else {
			tx.Commit()
			c.JSON(http.StatusCreated, auth)
		}
	}
}
