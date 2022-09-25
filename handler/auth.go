// Package handler ...
package handler

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/service"
)

// CreateUserAuth handles tasks for controller.CreateUserAuth
func CreateUserAuth(auth model.Auth) (httpResponse model.HTTPResponse, httpStatusCode int) {
	db := database.GetDB()
	authFinal := new(model.Auth)

	// email validation
	if !lib.ValidateEmail(auth.Email) {
		httpResponse.Message = "wrong email address"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// email must be unique
	if err := db.Where("email = ?", auth.Email).First(&auth).Error; err == nil {
		httpResponse.Message = "email already registered"
		httpStatusCode = http.StatusForbidden
		return
	}

	// user must not be able to manipulate all fields
	authFinal.Email = auth.Email
	authFinal.Password = auth.Password

	if service.SendEmail(authFinal.Email, model.EmailTypeVerification) {
		authFinal.VerifyEmail = model.EmailNotVerified
	}

	// one unique email for each account
	tx := db.Begin()
	if err := tx.Create(&authFinal).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1001")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	httpResponse.Message = *authFinal
	httpStatusCode = http.StatusCreated
	return
}
