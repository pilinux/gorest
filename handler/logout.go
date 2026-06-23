package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/mediocregopher/radix/v4"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
)

// Logout invalidates access and refresh tokens.
//
// When Redis is enabled, it stores the token JTIs in a blacklist with TTLs based
// on the token expirations. When Redis is disabled, it returns success without
// server-side invalidation.
func Logout(jtiAccess, jtiRefresh string, expAccess, expRefresh int64) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// Redis not enabled
	if !config.IsRedis() {
		httpResponse.Message = "logout successful"
		httpStatusCode = http.StatusOK
		return
	}

	// Redis enabled
	client := database.GetRedis()
	if client == nil {
		log.Error("error code: 1016.0: redis client not initialized")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	rConnTTL := config.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	jwtAccess := model.KeyValue{}
	jwtRefresh := model.KeyValue{}

	now := time.Now().Unix()

	// only blacklist tokens that are still valid; an already-expired token cannot
	// be used and writing it with a past EXAT would drop the key immediately
	if len(jtiAccess) > 0 && expAccess > now {
		jwtAccess.Key = config.PrefixJtiBlacklist + jtiAccess
		jwtAccess.Value = strconv.FormatInt(expAccess, 10)

		// set key and TTL atomically to avoid a TTL-less entry on crash
		if err := client.Do(ctx, radix.FlatCmd(nil, "SET", jwtAccess.Key, jwtAccess.Value, "EXAT", jwtAccess.Value)); err != nil {
			log.WithError(err).Error("error code: 1016.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}

	if len(jtiRefresh) > 0 && expRefresh > now {
		jwtRefresh.Key = config.PrefixJtiBlacklist + jtiRefresh
		jwtRefresh.Value = strconv.FormatInt(expRefresh, 10)

		// set key and TTL atomically to avoid a TTL-less entry on crash
		if err := client.Do(ctx, radix.FlatCmd(nil, "SET", jwtRefresh.Key, jwtRefresh.Value, "EXAT", jwtRefresh.Value)); err != nil {
			log.WithError(err).Error("error code: 1016.3")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}

	httpResponse.Message = "logout successful"
	httpStatusCode = http.StatusOK
	return
}
