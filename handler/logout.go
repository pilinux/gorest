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

// Logout handles jobs for controller.Logout
func Logout(jtiAccess, jtiRefresh string, expAccess, expRefresh int64) (httpResponse model.HTTPResponse, httpStatusCode int) {
	// Redis not enabled
	if !config.IsRedis() {
		httpResponse.Message = "logout successful"
		httpStatusCode = http.StatusOK
		return
	}

	// Redis enabled
	client := *database.GetRedis()
	rConnTTL := config.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	jwtAccess := model.KeyValue{}
	jwtRefresh := model.KeyValue{}

	if len(jtiAccess) > 0 {
		jwtAccess.Key = config.PrefixJtiBlacklist + jtiAccess
		jwtAccess.Value = strconv.FormatInt(expAccess, 10)

		// set key in Redis
		if err := client.Do(ctx, radix.FlatCmd(nil, "SET", jwtAccess.Key, jwtAccess.Value)); err != nil {
			log.WithError(err).Error("error code: 1016.1")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// set TTL
		if err := client.Do(ctx, radix.FlatCmd(nil, "EXPIREAT", jwtAccess.Key, expAccess)); err != nil {
			log.WithError(err).Error("error code: 1016.2")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}

	if len(jtiRefresh) > 0 {
		jwtRefresh.Key = config.PrefixJtiBlacklist + jtiRefresh
		jwtRefresh.Value = strconv.FormatInt(expRefresh, 10)

		// set key in Redis
		if err := client.Do(ctx, radix.FlatCmd(nil, "SET", jwtRefresh.Key, jwtRefresh.Value)); err != nil {
			log.WithError(err).Error("error code: 1016.3")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}

		// set TTL
		if err := client.Do(ctx, radix.FlatCmd(nil, "EXPIREAT", jwtRefresh.Key, expRefresh)); err != nil {
			log.WithError(err).Error("error code: 1016.4")
			httpResponse.Message = "internal server error"
			httpStatusCode = http.StatusInternalServerError
			return
		}
	}

	httpResponse.Message = "logout successful"
	httpStatusCode = http.StatusOK
	return
}
