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
var globalSentryClient *sentry.Client  // client backing the global hook

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
	// Capture the issue before delegating to the log hook. When this hook is
	// attached to the global logger, logrus core exits on Fatal and panics on
	// Panic right after all hooks fire, so the entry must be captured up front to
	// avoid losing it.
	h.captureIssue(entry)

	// Panic (0) and Fatal (1) lead to the process being torn down (by logrus
	// core, not the log hook); flush so the captured issue is delivered first.
	if entry.Level <= log.FatalLevel {
		h.Flush(2 * time.Second)
	}

	return h.logHook.Fire(entry)
}

func (h *sentryCombinedHook) Flush(timeout time.Duration) bool {
	return h.logHook.Flush(timeout)
}

func (h *sentryCombinedHook) FlushWithContext(ctx context.Context) bool {
	return h.logHook.FlushWithContext(ctx)
}

// scopeFromEntry returns the scope to capture the entry with. When the log
// entry carries a request-scoped hub (set by SentryCapture and propagated via
// log.WithContext(c.Request.Context())), its scope is cloned so the per-request
// tags and request data are included. Otherwise a clean scope is returned so a
// plain log.Error call is still captured, just without request enrichment.
//
// Note it never reads sentry.CurrentHub(): this service never calls
// sentry.Init, so the current hub has no client and an empty scope. Capture is
// always driven by the hook's own client (see captureIssue), so a missing
// request context only drops enrichment, never the event itself.
func scopeFromEntry(entry *log.Entry) *sentry.Scope {
	if entry.Context != nil {
		if hub := sentry.GetHubFromContext(entry.Context); hub != nil && hub.Scope() != nil {
			return hub.Scope().Clone()
		}
	}
	return sentry.NewScope()
}

func (h *sentryCombinedHook) captureIssue(entry *log.Entry) {
	if h.client == nil {
		return
	}

	// Every log level is captured as a Sentry issue (Fire runs this before the
	// terminating log hook so Fatal/Panic entries are captured too). Capturing
	// goes through a hub bound to this hook's own client so delivery never
	// depends on the request context or on a globally-initialized hub.
	hub := sentry.NewHub(h.client, scopeFromEntry(entry))
	hub.Scope().SetLevel(sentryLevelMap[entry.Level])

	for k, v := range entry.Data {
		switch val := v.(type) {
		case string:
			hub.Scope().SetTag(k, val)
		default:
			hub.Scope().SetTag(k, fmt.Sprint(v))
		}
	}

	if errVal, ok := entry.Data[log.ErrorKey].(error); ok {
		hub.Scope().SetContext("logrus", sentry.Context{"message": entry.Message})
		hub.CaptureException(errVal)
		return
	}

	hub.CaptureMessage(entry.Message)
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
	if ch, ok := hook.(*sentryCombinedHook); ok {
		globalSentryClient = ch.client
	}

	// Set log level and formatter globally
	if len(v) > 0 {
		if v[0] == "production" {
			log.SetLevel(log.InfoLevel)
		} else {
			log.SetLevel(log.TraceLevel)
		}
	} else {
		log.SetLevel(log.TraceLevel)
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
	globalSentryClient = nil
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
		// SendDefaultPII is left disabled (default), so sentry-go already strips
		// cookies and sensitive headers (e.g. Authorization) from captured
		// requests. The query string is not gated by that flag, and in this
		// service it can carry secrets (password-reset, email-verification and
		// 2FA tokens), so it is scrubbed here before any event leaves the process.
		// Cookies are cleared again as a defensive belt-and-suspenders measure.
		BeforeSend: func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
			if event.Request != nil {
				event.Request.QueryString = ""
				event.Request.Cookies = ""
			}
			return event
		},
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
			WithContext(entry.Context).
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
					WithContext(c.Request.Context()).
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

			hub := sentry.CurrentHub().Clone()
			// When context (log.WithContext(...)) is missing, bind the
			// global client to the hub so events are still captured.
			if hub.Client() == nil && globalSentryClient != nil {
				hub.BindClient(globalSentryClient)
			}

			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetTags(map[string]string{
					"method":      c.Request.Method,
					"path":        c.Request.URL.Path,
					"host":        c.Request.Host,
					"remote.addr": c.Request.RemoteAddr,
					"user.agent":  c.Request.UserAgent(),
				})
				scope.SetRequest(c.Request)
			})

			ctx := sentry.SetHubOnContext(c.Request.Context(), hub)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}
