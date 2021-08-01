package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/onrik/logrus/sentry"
	log "github.com/sirupsen/logrus"
)

// SentryCapture ...
func SentryCapture(sentryDsn string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Automatic recovery from panic
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic code: 100")
			}
		}()

		sentryHook, err := sentry.NewHook(sentry.Options{
			Dsn: sentryDsn,
		})
		if err != nil {
			log.WithError(err).Error("error code: 200")
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
