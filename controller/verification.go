package controller

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"

	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	"github.com/mediocregopher/radix/v4"
	log "github.com/sirupsen/logrus"
)

// VerifyEmail - verify email address
func VerifyEmail(c *gin.Context) {
	payload := struct {
		VerificationCode string `json:"verificationCode"`
	}{}

	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	data := struct {
		key   string
		value string
	}{}
	data.key = model.EmailVerificationKeyPrefix + payload.VerificationCode

	// get redis client
	client := *database.GetRedis()
	rConnTTL := config.GetConfig().Database.REDIS.Conn.ConnTTL

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// is key available in redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "EXISTS", data.key)); err != nil {
		log.WithError(err).Error("error code: 1061")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	if result == 0 {
		renderer.Render(c, gin.H{"msg": "wrong/expired verification code"}, http.StatusUnauthorized)
		return
	}

	// find key in redis
	if err := client.Do(ctx, radix.FlatCmd(&data.value, "GET", data.key)); err != nil {
		log.WithError(err).Error("error code: 1062")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	// delete key from redis
	result = 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "DEL", data.key)); err != nil {
		log.WithError(err).Error("error code: 1063")
	}
	if result == 0 {
		err := errors.New("failed to delete recovery key from redis")
		log.WithError(err).Error("error code: 1064")
	}

	// update verification status in database
	db := database.GetDB()
	auth := model.Auth{}

	if err := db.Where("email = ?", data.value).First(&auth).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "unknown user"}, http.StatusUnauthorized)
		return
	}

	if auth.VerifyEmail == model.EmailVerified {
		renderer.Render(c, gin.H{"msg": "email already verified"}, http.StatusOK)
		return
	}

	auth.VerifyEmail = model.EmailVerified
	auth.UpdatedAt = time.Now().Local()

	tx := db.Begin()
	if err := tx.Save(&auth).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1065")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	tx.Commit()

	renderer.Render(c, gin.H{"msg": "email successfully verified"}, http.StatusOK)
}

// CreateVerificationEmail issues new verification code upon request
func CreateVerificationEmail(c *gin.Context) {
	var payload LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	payload.Email = strings.TrimSpace(payload.Email)
	if !lib.ValidateEmail(payload.Email) {
		renderer.Render(c, gin.H{"msg": "wrong email address"}, http.StatusBadRequest)
		return
	}

	v, err := service.GetUserByEmail(payload.Email)
	if err != nil {
		renderer.Render(c, gin.H{"msg": "user not found"}, http.StatusNotFound)
		return
	}

	// is email already verified
	if v.VerifyEmail == model.EmailVerified {
		renderer.Render(c, gin.H{"msg": "email already verified"}, http.StatusOK)
		return
	}

	// verify password
	verifyPass, err := argon2id.ComparePasswordAndHash(payload.Password, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1071")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	if !verifyPass {
		renderer.Render(c, gin.H{"msg": "wrong credentials"}, http.StatusUnauthorized)
		return
	}

	// issue new verification code
	if !sendEmail(v.Email, model.EmailTypeVerification) {
		renderer.Render(c, gin.H{"msg": "failed to send verification email"}, http.StatusServiceUnavailable)
		return
	}

	renderer.Render(c, gin.H{"msg": "sent verification email"}, http.StatusOK)
}
