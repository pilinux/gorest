package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
)

// RateLimit - rate limit middleware
func RateLimit(limiterInstance *limiter.Limiter) gin.HandlerFunc {
	// limiter instance is nil
	if limiterInstance == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// give the limiter instance to the middleware initializer
	return mgin.NewMiddleware(limiterInstance)
}
