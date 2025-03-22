package middleware

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import (
	"net/http"
	"strconv"
	"time"

	sentry "github.com/getsentry/sentry-go"
	sentrylogrus "github.com/getsentry/sentry-go/logrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var globalSentryHook *sentrylogrus.Hook // global hook

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
func InitSentry(sentryDsn string, v ...string) (*sentrylogrus.Hook, error) {
	if globalSentryHook != nil {
		// prevent double hook
		return globalSentryHook, nil
	}

	hook, err := createSentryHook(sentryDsn, v...)
	if err != nil {
		return nil, err
	}

	// Attach the hook only once globally
	log.AddHook(hook)
	globalSentryHook = hook

	// Set log level and formatter globally
	if len(v) > 0 {
		if v[0] == "production" {
			log.SetLevel(log.InfoLevel)
		} else {
			log.SetLevel(log.DebugLevel)
		}
	} else {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(&log.JSONFormatter{})

	return globalSentryHook, nil
}

// NewSentryHook creates a new Sentry hook for goroutine-specific loggers
func NewSentryHook(sentryDsn string, v ...string) (*sentrylogrus.Hook, error) {
	return createSentryHook(sentryDsn, v...)
}

// DestroySentry - destroy global sentry hook
func DestroySentry() {
	if globalSentryHook != nil {
		globalSentryHook.Flush(5 * time.Second)
		globalSentryHook = nil
	}
}

func createSentryHook(sentryDsn string, v ...string) (*sentrylogrus.Hook, error) {
	sentryDebugMode := true
	environment := "development" // default
	release := ""
	enableTracing := false
	tracesSampleRate := 0.0
	if len(v) > 0 {
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

	// Sentry Client options
	clientOptions := sentry.ClientOptions{
		Dsn:              sentryDsn,
		Debug:            sentryDebugMode,
		Environment:      environment,
		Release:          release,
		EnableTracing:    enableTracing,
		TracesSampleRate: tracesSampleRate,
	}

	// Hook with desired log levels
	hook, err := sentrylogrus.New(log.AllLevels, clientOptions)
	if err != nil {
		return nil, err
	}

	// Add fallback behavior if Sentry fails
	hook.SetFallback(func(entry *log.Entry) error {
		// Log the failure locally
		log.
			WithFields(entry.Data).
			Error("Sentry fallback executed")

		return nil
	})

	return hook, nil
}

// SentryCapture - sentry middleware to capture errors and forward to sentry.io
func SentryCapture() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Automatic recovery from panic
		defer func() {
			if r := recover(); r != nil {
				log.
					WithField("panic", r).
					Error("panic msg: middleware -> sentry panicked")

				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}()

		if globalSentryHook != nil {
			// Ensure flushing Sentry before responding
			defer func() {
				globalSentryHook.Flush(5 * time.Second)
			}()

			// Enrich the Sentry Scope with request context
			globalSentryHook.AddTags(map[string]string{
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
				"host":        c.Request.Host,
				"remote.addr": c.Request.RemoteAddr,
				"user.agent":  c.Request.UserAgent(),
			})
		}

		c.Next()
	}
}
