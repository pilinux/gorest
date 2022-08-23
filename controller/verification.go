package controller

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib/renderer"

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
	data.key = "gorest-email-verification-" + payload.VerificationCode

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
		err := errors.New("failed to delete key from redis: " + data.key)
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
