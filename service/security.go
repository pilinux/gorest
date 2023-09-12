package service

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mediocregopher/radix/v4"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
)

// IsTokenAllowed returns true when the token is not in the blacklist
//
// Dependency: JWT, Redis database + enable 'INVALIDATE_JWT' in .env
func IsTokenAllowed(jti string) bool {
	// verify that JWT service is enabled in .env
	if !config.IsJWT() {
		return true
	}

	// Redis not available, abort
	if !config.IsRedis() {
		return true
	}

	// token blacklist management not enabled, abort
	if !config.InvalidateJWT() {
		return true
	}

	jti = config.PrefixJtiBlacklist + jti

	client := *database.GetRedis()
	rConnTTL := config.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// is key available in Redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "EXISTS", jti)); err != nil {
		log.WithError(err).Error("error code: 501")
		return false
	}

	// key found in blacklist
	if result != 0 {
		return false
	}

	// key not found in blacklist
	return true
}

// JWTBlacklistChecker validates a token against the blacklist
func JWTBlacklistChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		var jti string
		jtiAccess := strings.TrimSpace(c.GetString("jtiAccess"))
		jtiRefresh := strings.TrimSpace(c.GetString("jtiRefresh"))

		if jtiAccess != "" {
			jti = jtiAccess
			goto CheckBlackList
		}
		if jtiRefresh != "" {
			jti = jtiRefresh
			goto CheckBlackList
		}

	CheckBlackList:
		if !IsTokenAllowed(jti) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "invalid token")
			return
		}

		c.Next()
	}
}
