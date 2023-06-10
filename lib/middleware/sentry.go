package middleware

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/onrik/logrus/sentry"
	log "github.com/sirupsen/logrus"
)

// SentryCapture - capture errors and forward to sentry.io
//
// required parameter (1st parameter): sentryDsn
//
// optional parameter (2nd parameter): environment (development or production)
//
// optional parameter (3rd parameter): release version or git commit number
func SentryCapture(sentryDsn string, v ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Automatic recovery from panic
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic msg: middleware -> sentry panicked")
			}
		}()

		sentryDebugMode := true
		release := ""
		if len(v) >= 1 {
			if v[0] == "production" {
				sentryDebugMode = false
			}
		}
		if len(v) > 1 {
			release = strings.TrimSpace(v[1])
		}

		sentryHook, err := sentry.NewHook(sentry.Options{
			Dsn:     sentryDsn,
			Debug:   sentryDebugMode,
			Release: release,
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
