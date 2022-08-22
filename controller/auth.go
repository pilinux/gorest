// Package controller contains all the controllers
// of the application
package controller

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/mediocregopher/radix/v4"
	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib"
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
	if !lib.ValidateEmail(auth.Email) {
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

	if sendVerificationEmail(authFinal.Email) {
		authFinal.VerifyEmail = model.EmailNotVerified
	}

	// one unique email for each account
	tx := db.Begin()
	if err := tx.Create(&authFinal).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1001")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	tx.Commit()

	renderer.Render(c, authFinal, http.StatusCreated)
}

// sendVerificationEmail sends a verification email if
// - required by the application
// - an external email service is configured
// - a redis database is configured
func sendVerificationEmail(email string) bool {
	// send verification email if required by the application
	appConfig := config.GetConfig()
	// is external email service activated
	if appConfig.EmailConf.Activate == config.Activated {
		// is email verification required
		if appConfig.Security.VerifyEmail {
			// is redis database activated
			if appConfig.Database.REDIS.Activate == config.Activated {
				// generate verification code
				verificationCode := lib.SecureRandomNumber(uint(appConfig.EmailConf.EmailVerificationCodeLength))

				data := struct {
					key   string
					value string
				}{}
				data.key = "gorest-email-verification-" + strconv.FormatUint(verificationCode, 10)
				data.value = email

				// save in redis with expiry time
				client := *database.GetRedis()
				redisConnTTL := appConfig.Database.REDIS.Conn.ConnTTL

				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(redisConnTTL)*time.Second)
				defer cancel()

				// Set key in Redis
				r1 := ""
				if err := client.Do(ctx, radix.FlatCmd(&r1, "SET", data.key, data.value)); err != nil {
					log.WithError(err).Error("error code: 1002")
				}
				if r1 != "OK" {
					log.Error("error code: 1003")
				}

				// Set expiry time
				r2 := 0
				if err := client.Do(ctx, radix.FlatCmd(&r2, "EXPIRE", data.key, appConfig.EmailConf.EmailValidityPeriod)); err != nil {
					log.WithError(err).Error("error code: 1004")
				}
				if r2 != 1 {
					log.Error("error code: 1005")
				}

				// check which email service
				// for Postmark
				if appConfig.EmailConf.Provider == "postmark" {
					htmlModel := lib.HTMLModel(lib.StrArrHTMLModel(appConfig.EmailConf.HTMLModel))
					htmlModel["verification_code"] = verificationCode

					params := service.PostmarkParams{}
					params.ServerToken = appConfig.EmailConf.APIToken
					params.TemplateID = appConfig.EmailConf.EmailVerificationTemplateID
					params.From = appConfig.EmailConf.AddrFrom
					params.To = email
					params.Tag = appConfig.EmailConf.EmailVerificationTag
					params.TrackOpens = appConfig.EmailConf.TrackOpens
					params.TrackLinks = appConfig.EmailConf.TrackLinks
					params.MessageStream = appConfig.EmailConf.MsgType
					params.HTMLModel = htmlModel

					// send the email
					res, err := service.Postmark(params)
					if err != nil {
						log.WithError(err).Error("error code: 1006")
					}
					if res.Message != "OK" {
						log.Error(res)
					}
				}
				return true
			}
		}
	}
	return false
}
