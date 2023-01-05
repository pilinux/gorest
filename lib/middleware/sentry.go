package middleware

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/onrik/logrus/sentry"
	log "github.com/sirupsen/logrus"
)

// SentryCapture - capture errors and forward to sentry.io
func SentryCapture(sentryDsn string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Automatic recovery from panic
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic msg: middleware -> sentry panicked")
			}
		}()

		sentryHook, err := sentry.NewHook(sentry.Options{
			Dsn: sentryDsn,
		})
		if err != nil {
			// middleware -> sentry NewHook failed
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		sentryHook.AddTag("method", c.Request.Method)
		sentryHook.AddTag("path", c.Request.URL.Path)
		sentryHook.AddTag("host", c.Request.Host)
		sentryHook.AddTag("remote.addr", c.Request.RemoteAddr)
		sentryHook.AddTag("user.agent", c.Request.UserAgent())
		defer sentryHook.Flush()

		log.AddHook(sentryHook)

		c.Next()
	}
}
