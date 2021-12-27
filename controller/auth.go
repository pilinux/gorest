package controller

import (
	"net/http"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/service"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// CreateUserAuth - POST /register
func CreateUserAuth(c *gin.Context) {
	db := database.GetDB()
	auth := model.Auth{}
	createAuth := 0 // default value

	if err := c.ShouldBindJSON(&auth); err != nil {
		render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	if !service.IsEmailValid(auth.Email) {
		createAuth = 1 // invalid email
		render(c, gin.H{"msg": "wrong email address"}, http.StatusBadRequest)
		return
	}

	if err := db.Where("email = ?", auth.Email).First(&auth).Error; err == nil {
		createAuth = 2 // email is already registered
		render(c, gin.H{"msg": "email already registered"}, http.StatusBadRequest)
		return
	}

	// one unique email for each account
	if createAuth == 0 {
		tx := db.Begin()
		if err := tx.Create(&auth).Error; err != nil {
			tx.Rollback()
			log.WithError(err).Error("error code: 1001")
			render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		} else {
			tx.Commit()
			render(c, auth, http.StatusCreated)
		}
	}
}
