package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mediocregopher/radix/v4"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/lib/renderer"
)

// RedisData - key:value
type RedisData struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// RedisDataHash - key:hashValues
type RedisDataHash struct {
	Key   string `json:"Key"`
	Value RedisDataHashValue
}

// RedisDataHashValue - values
type RedisDataHashValue struct {
	Value1 string `json:"Value1"`
	Value2 string `json:"Value2"`
	Value3 string `json:"Value3"`
	Value4 string `json:"Value4"`
}

// RedisCreate - SET key
func RedisCreate(c *gin.Context) {
	data := RedisData{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	client := database.GetRedis()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.RedisConnTTL)*time.Second)
	defer cancel()

	// Set key in Redis
	result := ""
	if err := client.Do(ctx, radix.FlatCmd(&result, "SET", data.Key, data.Value)); err != nil {
		log.WithError(err).Error("error code: 1301")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	if result != "OK" {
		renderer.Render(c, gin.H{"msg": "operation failed"}, http.StatusNotAcceptable)
		return
	}

	renderer.Render(c, data, http.StatusOK)
}

// RedisRead - GET key
func RedisRead(c *gin.Context) {
	data := RedisData{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	client := database.GetRedis()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.RedisConnTTL)*time.Second)
	defer cancel()

	// Is key available in Redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "EXISTS", data.Key)); err != nil {
		log.WithError(err).Error("error code: 1311")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	if result == 0 {
		renderer.Render(c, gin.H{"msg": "key does not exist"}, http.StatusNotFound)
		return
	}

	// Find key in Redis
	if err := client.Do(ctx, radix.FlatCmd(&data.Value, "GET", data.Key)); err != nil {
		log.WithError(err).Error("error code: 1312")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	renderer.Render(c, data, http.StatusOK)
}

// RedisDelete - DEL key
func RedisDelete(c *gin.Context) {
	data := RedisData{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	client := database.GetRedis()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.RedisConnTTL)*time.Second)
	defer cancel()

	// Delete key in Redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "DEL", data.Key)); err != nil {
		log.WithError(err).Error("error code: 1321")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	if result == 0 {
		renderer.Render(c, gin.H{"msg": "key does not exist"}, http.StatusNotFound)
		return
	}

	renderer.Render(c, gin.H{"msg": "key is deleted"}, http.StatusOK)
}

// RedisCreateHash - SET hashes
func RedisCreateHash(c *gin.Context) {
	data := RedisDataHash{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	client := database.GetRedis()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.RedisConnTTL)*time.Second)
	defer cancel()

	// Set key in Redis
	if err := client.Do(ctx, radix.FlatCmd(nil, "HSET", data.Key, data.Value)); err != nil {
		log.WithError(err).Error("error code: 1331")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	renderer.Render(c, data, http.StatusOK)
}

// RedisReadHash - GET hashes
func RedisReadHash(c *gin.Context) {
	data := RedisDataHash{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	client := database.GetRedis()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.RedisConnTTL)*time.Second)
	defer cancel()

	// Is key available in Redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "EXISTS", data.Key)); err != nil {
		log.WithError(err).Error("error code: 1341")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	if result == 0 {
		renderer.Render(c, gin.H{"msg": "key does not exist"}, http.StatusNotFound)
		return
	}

	// Find key in Redis
	if err := client.Do(ctx, radix.FlatCmd(&data.Value, "HGETALL", data.Key)); err != nil {
		log.WithError(err).Error("error code: 1342")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	renderer.Render(c, data, http.StatusOK)
}

// RedisDeleteHash - DEL hashes
func RedisDeleteHash(c *gin.Context) {
	data := RedisDataHash{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	client := database.GetRedis()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.RedisConnTTL)*time.Second)
	defer cancel()

	// Delete key in Redis
	result := 0
	if err := client.Do(ctx, radix.FlatCmd(&result, "HDEL", data.Key, data.Value)); err != nil {
		log.WithError(err).Error("error code: 1351")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	if result == 0 {
		renderer.Render(c, gin.H{"msg": "key does not exist"}, http.StatusNotFound)
		return
	}

	renderer.Render(c, gin.H{"msg": "key is deleted"}, http.StatusOK)
}

// AccessResource - can be accessed by basic auth
func AccessResource(c *gin.Context) {
	renderer.Render(c, gin.H{"msg": "access granted!"}, http.StatusOK)
}
