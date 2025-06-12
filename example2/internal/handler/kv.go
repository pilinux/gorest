package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	gconfig "github.com/pilinux/gorest/config"
	grenderer "github.com/pilinux/gorest/lib/renderer"

	"github.com/pilinux/gorest/example2/internal/database/model"
	"github.com/pilinux/gorest/example2/internal/service"
)

// KeyValueAPI provides HTTP handlers for Redis key-value store endpoints.
type KeyValueAPI struct {
	keyValueService *service.KeyValueService
}

// NewKeyValueAPI returns a new KeyValueAPI instance.
func NewKeyValueAPI(keyValueService *service.KeyValueService) *KeyValueAPI {
	return &KeyValueAPI{
		keyValueService: keyValueService,
	}
}

// SetKeyValue handles the HTTP POST request to set a key-value pair in Redis.
//
// Endpoint: POST /api/v1/kv
//
// Authorization: None
func (api *KeyValueAPI) SetKeyValue(c *gin.Context) {
	data := model.KV{}
	if err := c.ShouldBindJSON(&data); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	if err := data.Trim(); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	rConnTTL := gconfig.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	res, statusCode := api.keyValueService.SetKeyValue(ctx, &data)
	grenderer.Render(c, res, statusCode)
}

// GetKeyValue handles the HTTP GET request to retrieve a value by key from Redis.
//
// Endpoint: GET /api/v1/kv/:key
//
// Authorization: None
func (api *KeyValueAPI) GetKeyValue(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		grenderer.Render(c, gin.H{"message": "key is required"}, http.StatusBadRequest)
		return
	}

	rConnTTL := gconfig.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	res, statusCode := api.keyValueService.GetKeyValue(ctx, key)
	grenderer.Render(c, res, statusCode)
}

// DeleteKeyValue handles the HTTP DELETE request to delete a key-value pair from Redis.
//
// Endpoint: DELETE /api/v1/kv/:key
//
// Authorization: None
func (api *KeyValueAPI) DeleteKeyValue(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		grenderer.Render(c, gin.H{"message": "key is required"}, http.StatusBadRequest)
		return
	}

	rConnTTL := gconfig.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	res, statusCode := api.keyValueService.DeleteKeyValue(ctx, key)
	grenderer.Render(c, res, statusCode)
}

// SetHashKeyValue handles the HTTP POST request to set a hash key-value pair in Redis.
func (api *KeyValueAPI) SetHashKeyValue(c *gin.Context) {
	data := model.KVHash{}
	if err := c.ShouldBindJSON(&data); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	if err := data.Trim(); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	rConnTTL := gconfig.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	res, statusCode := api.keyValueService.SetHashKeyValue(ctx, &data)
	grenderer.Render(c, res, statusCode)
}

// GetHashKeyValue handles the HTTP GET request to retrieve a hash value by key from Redis.
func (api *KeyValueAPI) GetHashKeyValue(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		grenderer.Render(c, gin.H{"message": "key is required"}, http.StatusBadRequest)
		return
	}

	rConnTTL := gconfig.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	res, statusCode := api.keyValueService.GetHashKeyValue(ctx, key)
	grenderer.Render(c, res, statusCode)
}

// DeleteHashKeyValue handles the HTTP DELETE request to delete a hash key-value pair from Redis.
func (api *KeyValueAPI) DeleteHashKeyValue(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		grenderer.Render(c, gin.H{"message": "key is required"}, http.StatusBadRequest)
		return
	}

	rConnTTL := gconfig.GetConfig().Database.REDIS.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
	defer cancel()

	res, statusCode := api.keyValueService.DeleteHashKeyValue(ctx, key)
	grenderer.Render(c, res, statusCode)
}
