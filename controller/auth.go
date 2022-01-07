package controller

import (
	"net/http"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// CreateUserAuth - POST /register
func CreateUserAuth(c *gin.Context) {
	db := database.GetDB()
	auth := model.Auth{}
	authFinal := model.Auth{}

	// bind JSON
	if err := c.ShouldBindJSON(&auth); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	// email validation
	if !service.IsEmailValid(auth.Email) {
		renderer.Render(c, gin.H{"msg": "wrong email address"}, http.StatusBadRequest)
		return
	}

	// email must be unique
	if err := db.Where("email = ?", auth.Email).First(&auth).Error; err == nil {
		renderer.Render(c, gin.H{"msg": "email already registered"}, http.StatusForbidden)
		return
	}

	// user must not be able to manipulate all fields
	authFinal.Email = auth.Email
	authFinal.Password = auth.Password

	// one unique email for each account
	tx := db.Begin()
	if err := tx.Create(&authFinal).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1001")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
	} else {
		tx.Commit()
		renderer.Render(c, authFinal, http.StatusCreated)
	}
}
