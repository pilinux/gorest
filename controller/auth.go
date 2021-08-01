package controller

import (
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// CreateUserAuth - POST /register
func CreateUserAuth(c *gin.Context) {
	db := database.GetDB()
	auth := model.Auth{}
	createAuth := 0 // default value

	c.ShouldBindJSON(&auth)

	if !isEmailValid(auth.Email) {
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

// isEmailValid checks if the email provided passes the required structure
// and length test. It also checks the domain has a valid MX record.
// Credit: Edd Turtle
func isEmailValid(e string) bool {
	if len(e) < 3 || len(e) > 254 {
		return false
	}

	if !emailRegex.MatchString(e) {
		return false
	}

	parts := strings.Split(e, "@")
	mx, err := net.LookupMX(parts[1])
	if err != nil || len(mx) == 0 {
		return false
	}

	return true
}
