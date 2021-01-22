package controller

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/piLinux/GoREST/database"
	"github.com/piLinux/GoREST/database/model"

	"github.com/gin-gonic/gin"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// CreateUserAuth - POST /register
func CreateUserAuth(c *gin.Context) {
	db := database.GetDB()
	auth := model.Auth{}
	createAuth := 0 // default value

	c.ShouldBindJSON(&auth)

	if isEmailValid(auth.Email) == false {
		createAuth = 1 // invalid email
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := db.Where("email = ?", auth.Email).First(&auth).Error; err == nil {
		createAuth = 2 // email is already registered
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
