package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/mediocregopher/radix/v4"
	log "github.com/sirupsen/logrus"

	gconfig "github.com/pilinux/gorest/config"
	gdatabase "github.com/pilinux/gorest/database"
	gmodel "github.com/pilinux/gorest/database/model"

	"github.com/pilinux/gorest/example/database/model"
)

// RedisCreate - handles jobs for controller.RedisCreate
func RedisCreate(data model.RedisData) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	client := *gdatabase.GetRedis()
	rConnTTL := gconfig.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// set key in Redis
	result := ""
	if err := client.Do(ctx, radix.FlatCmd(&result, "SET", data.Key, data.Value)); err != nil {
		log.WithError(err).Error("error code: 1301")
		httpResponse.Result = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if result != "OK" {
		httpResponse.Result = "operation failed"
		httpStatusCode = http.StatusNotAcceptable
		return
	}

	httpResponse.Result = data
	httpStatusCode = http.StatusCreated
	return
}

// RedisRead - handles jobs for controller.RedisRead
func RedisRead(data model.RedisData) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	client := *gdatabase.GetRedis()
	rConnTTL := gconfig.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// is key available in Redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "EXISTS", data.Key)); err != nil {
		log.WithError(err).Error("error code: 1311")
		httpResponse.Result = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if result == 0 {
		httpResponse.Result = "key does not exist"
		httpStatusCode = http.StatusNotFound
		return
	}

	// find key in Redis
	if err := client.Do(ctx, radix.FlatCmd(&data.Value, "GET", data.Key)); err != nil {
		log.WithError(err).Error("error code: 1312")
		httpResponse.Result = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Result = data
	httpStatusCode = http.StatusOK
	return
}

// RedisDelete - handles jobs for controller.RedisDelete
func RedisDelete(data model.RedisData) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	client := *gdatabase.GetRedis()
	rConnTTL := gconfig.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// delete key in Redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "DEL", data.Key)); err != nil {
		log.WithError(err).Error("error code: 1321")
		httpResponse.Result = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if result == 0 {
		httpResponse.Result = "key does not exist"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Result = "key is deleted"
	httpStatusCode = http.StatusOK
	return
}

// RedisCreateHash - handles jobs for controller.RedisCreateHash
func RedisCreateHash(data model.RedisDataHash) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	client := *gdatabase.GetRedis()
	rConnTTL := gconfig.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// set key in Redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "HSET", data.Key, data.Value)); err != nil {
		log.WithError(err).Error("error code: 1331")
		httpResponse.Result = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if result != 1 {
		httpResponse.Result = "operation failed"
		httpStatusCode = http.StatusNotAcceptable
		return
	}

	httpResponse.Result = data
	httpStatusCode = http.StatusCreated
	return
}

// RedisReadHash - handles jobs for controller.RedisReadHash
func RedisReadHash(data model.RedisDataHash) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	client := *gdatabase.GetRedis()
	rConnTTL := gconfig.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// is key available in Redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "EXISTS", data.Key)); err != nil {
		log.WithError(err).Error("error code: 1341")
		httpResponse.Result = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if result == 0 {
		httpResponse.Result = "key does not exist"
		httpStatusCode = http.StatusNotFound
		return
	}

	// find key in Redis
	if err := client.Do(ctx, radix.FlatCmd(&data.Value, "HGETALL", data.Key)); err != nil {
		log.WithError(err).Error("error code: 1342")
		httpResponse.Result = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Result = data
	httpStatusCode = http.StatusOK
	return
}

// RedisDeleteHash - handles jobs for controller.RedisDeleteHash
func RedisDeleteHash(data model.RedisDataHash) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	client := *gdatabase.GetRedis()
	rConnTTL := gconfig.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	// delete key in Redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "HDEL", data.Key, data.Value)); err != nil {
		log.WithError(err).Error("error code: 1351")
		httpResponse.Result = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if result == 0 {
		httpResponse.Result = "key does not exist"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Result = "key is deleted"
	httpStatusCode = http.StatusOK
	return
}
