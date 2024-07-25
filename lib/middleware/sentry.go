package middleware

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/logrus/sentry"
	log "github.com/sirupsen/logrus"
)

// InitSentry - initialize sentry for middleware or separate goroutines
//
// required parameter (1st parameter): sentryDsn
//
// optional parameter (2nd parameter): environment (development or production)
//
// optional parameter (3rd parameter): release version or git commit number
//
// optional parameter (4th parameter): enableTracing (yes or no)
//
// optional parameter (5th parameter): tracesSampleRate (0.0 - 1.0)
func InitSentry(sentryDsn string, v ...string) (*sentry.Hook, error) {
	sentryDebugMode := true
	environment := "development" // default
	release := ""
	enableTracing := false
	tracesSampleRate := 0.0
	if len(v) >= 1 {
		environment = v[0]
		if environment == "production" {
			sentryDebugMode = false
		}
	}
	if len(v) > 1 {
		release = v[1]
	}
	if len(v) > 2 {
		if v[2] == "yes" {
			enableTracing = true
		}
	}
	if len(v) > 3 {
		if enableTracing {
			sampleRate, err := strconv.ParseFloat(v[3], 64)
			if err == nil {
				tracesSampleRate = sampleRate
			}
		}
	}

	return sentry.NewHook(sentry.Options{
		Dsn:              sentryDsn,
		Debug:            sentryDebugMode,
		Environment:      environment,
		Release:          release,
		EnableTracing:    enableTracing,
		TracesSampleRate: tracesSampleRate,
	})
}

// SentryCapture - sentry middleware to capture errors and forward to sentry.io
//
// required parameter (1st parameter): sentryDsn
//
// optional parameter (2nd parameter): environment (development or production)
//
// optional parameter (3rd parameter): release version or git commit number
//
// optional parameter (4th parameter): enableTracing (yes or no)
//
// optional parameter (5th parameter): tracesSampleRate (0.0 - 1.0)
func SentryCapture(sentryDsn string, v ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Automatic recovery from panic
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic msg: middleware -> sentry panicked")
			}
		}()

		sentryHook, err := InitSentry(sentryDsn, v...)
		if err != nil {
			// middleware -> sentry NewHook failed
			c.AbortWithStatusJSON(http.StatusInternalServerError, "internal server error")
			return
		}
		sentryHook.AddTag("method", c.Request.Method)
		sentryHook.AddTag("path", c.Request.URL.Path)
		sentryHook.AddTag("host", c.Request.Host)
		sentryHook.AddTag("remote.addr", c.Request.RemoteAddr)
		sentryHook.AddTag("user.agent", c.Request.UserAgent())
		sentryHook.SetFlushTimeout(5 * time.Second)
		defer sentryHook.Flush()

		log.AddHook(sentryHook)

		c.Next()
	}
}
