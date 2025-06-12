package service

import (
	"context"
	"errors"
	"net/http"

	gmodel "github.com/pilinux/gorest/database/model"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/example2/internal/database/model"
	"github.com/pilinux/gorest/example2/internal/repo"
)

// KeyValueService provides methods to interact with Redis key-value store.
type KeyValueService struct {
	kvRepo repo.KeyValueRepository
}

// NewKeyValueService returns a new KeyValueService instance.
func NewKeyValueService(kvRepo repo.KeyValueRepository) *KeyValueService {
	return &KeyValueService{
		kvRepo: kvRepo,
	}
}

// SetKeyValue sets a key-value pair in Redis.
func (s *KeyValueService) SetKeyValue(ctx context.Context, kv *model.KV) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	result, err := s.kvRepo.SetKeyValue(ctx, kv.Key, kv.Value)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithError(err).Error("SetKeyValue.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if result != "OK" {
		httpResponse.Message = "operation failed"
		httpStatusCode = http.StatusNotAcceptable
		return
	}

	httpResponse.Message = kv
	httpStatusCode = http.StatusCreated
	return
}

// GetKeyValue retrieves a value by key from Redis.
func (s *KeyValueService) GetKeyValue(ctx context.Context, key string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	value, err := s.kvRepo.GetKeyValue(ctx, key)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithError(err).Error("GetKeyValue.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if value == "" {
		httpResponse.Message = "key not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = model.KV{Key: key, Value: value}
	httpStatusCode = http.StatusOK
	return
}

// DeleteKeyValue deletes a key-value pair from Redis.
func (s *KeyValueService) DeleteKeyValue(ctx context.Context, key string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	ok, err := s.kvRepo.DeleteKeyValue(ctx, key)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithError(err).Error("DeleteKeyValue.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if !ok {
		httpResponse.Message = "key not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = "key deleted successfully"
	httpStatusCode = http.StatusOK
	return
}

// SetHashKeyValue sets a hash key-value pair in Redis.
func (s *KeyValueService) SetHashKeyValue(ctx context.Context, kv *model.KVHash) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	err := s.kvRepo.SetHashKeyValue(ctx, kv.Key, kv.Value)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithError(err).Error("SetHashKeyValue.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = kv
	httpStatusCode = http.StatusCreated
	return
}

// GetHashKeyValue retrieves a hash value by key from Redis.
func (s *KeyValueService) GetHashKeyValue(ctx context.Context, key string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	value, err := s.kvRepo.GetHashKeyValue(ctx, key)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithError(err).Error("GetHashKeyValue.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if value == nil {
		httpResponse.Message = "hash key not found"
		httpStatusCode = http.StatusNotFound
		return
	}
	val, ok := value.(*model.KVHashValue)
	if !ok {
		httpResponse.Message = "invalid value type"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = model.KVHash{Key: key, Value: *val}
	httpStatusCode = http.StatusOK
	return
}

// DeleteHashKeyValue deletes a hash key-value pair from Redis.
func (s *KeyValueService) DeleteHashKeyValue(ctx context.Context, key string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	var value model.KVHashValue
	ok, err := s.kvRepo.DeleteHashKeyValue(ctx, key, &value)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithError(err).Error("DeleteHashKeyValue.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if !ok {
		httpResponse.Message = "hash key not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = "hash key deleted successfully"
	httpStatusCode = http.StatusOK
	return
}
