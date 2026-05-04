package middleware

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022-2026 pilinux

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	sentry "github.com/getsentry/sentry-go"
	sentrylogrus "github.com/getsentry/sentry-go/logrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var globalSentryHook sentrylogrus.Hook // global hook

var sentryLevelMap = map[log.Level]sentry.Level{
	log.TraceLevel: sentry.LevelDebug,
	log.DebugLevel: sentry.LevelDebug,
	log.InfoLevel:  sentry.LevelInfo,
	log.WarnLevel:  sentry.LevelWarning,
	log.ErrorLevel: sentry.LevelError,
	log.FatalLevel: sentry.LevelFatal,
	log.PanicLevel: sentry.LevelFatal,
}

type sentryCombinedHook struct {
	logHook sentrylogrus.Hook
	client  *sentry.Client
}

func (h *sentryCombinedHook) SetHubProvider(provider func() *sentry.Hub) {
	h.logHook.SetHubProvider(provider)
}

func (h *sentryCombinedHook) AddTags(tags map[string]string) {
	h.logHook.AddTags(tags)
}

func (h *sentryCombinedHook) SetFallback(fb sentrylogrus.FallbackFunc) {
	h.logHook.SetFallback(fb)
}

func (h *sentryCombinedHook) SetKey(oldKey, newKey string) {
	h.logHook.SetKey(oldKey, newKey)
}

func (h *sentryCombinedHook) Levels() []log.Level {
	return h.logHook.Levels()
}

func (h *sentryCombinedHook) Fire(entry *log.Entry) error {
	if err := h.logHook.Fire(entry); err != nil {
		return err
	}

	h.captureIssue(entry)
	return nil
}

func (h *sentryCombinedHook) Flush(timeout time.Duration) bool {
	return h.logHook.Flush(timeout)
}

func (h *sentryCombinedHook) FlushWithContext(ctx context.Context) bool {
	return h.logHook.FlushWithContext(ctx)
}

func (h *sentryCombinedHook) captureIssue(entry *log.Entry) {
	if h.client == nil {
		return
	}

	if entry.Level < log.ErrorLevel {
		return
	}

	scope := sentry.NewScope()
	scope.SetLevel(sentryLevelMap[entry.Level])

	for k, v := range entry.Data {
		switch val := v.(type) {
		case string:
			scope.SetTag(k, val)
		default:
			scope.SetTag(k, fmt.Sprint(v))
		}
	}

	if errVal, ok := entry.Data[log.ErrorKey].(error); ok {
		scope.SetContext("logrus", sentry.Context{"message": entry.Message})
		h.client.CaptureException(errVal, &sentry.EventHint{OriginalException: errVal}, scope)
		return
	}

	h.client.CaptureMessage(entry.Message, nil, scope)
}

// InitSentry initializes sentry for middleware or separate goroutines.
//
//   - required parameter (1st parameter): sentryDsn
//   - optional parameter (2nd parameter): environment (development or production)
//   - optional parameter (3rd parameter): release version or git commit number
//   - optional parameter (4th parameter): enableTracing (yes or no)
//   - optional parameter (5th parameter): tracesSampleRate (0.0 - 1.0)
func InitSentry(sentryDsn string, v ...string) (sentrylogrus.Hook, error) {
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

// NewSentryHook creates a new Sentry hook for goroutine-specific loggers.
func NewSentryHook(sentryDsn string, v ...string) (sentrylogrus.Hook, error) {
	return createSentryHook(sentryDsn, v...)
}

// DestroySentry destroys the global sentry hook.
func DestroySentry() {
	if globalSentryHook != nil {
		globalSentryHook.Flush(5 * time.Second)
		globalSentryHook = nil
	}
}

func createSentryHook(sentryDsn string, v ...string) (sentrylogrus.Hook, error) {
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
	// https://docs.sentry.io/platforms/go/configuration/options/
	// https://docs.sentry.io/platforms/go/tracing/trace-propagation/ (feature introduced in sentry-go v0.41.0)
	clientOptions := sentry.ClientOptions{
		Dsn:              sentryDsn,
		Debug:            sentryDebugMode,
		Environment:      environment,
		Release:          release,
		EnableTracing:    enableTracing,
		TracesSampleRate: tracesSampleRate,
		EnableLogs:       true, // enable logs to be captured
	}

	client, err := sentry.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}
	client.SetSDKIdentifier("sentry.go.logrus")

	logHook := sentrylogrus.NewLogHookFromClient(log.AllLevels, client)
	hook := &sentryCombinedHook{
		logHook: logHook,
		client:  client,
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

// SentryCapture returns a sentry middleware to capture errors and forward to sentry.io.
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
