package repo

import (
	"context"
	"reflect"

	"github.com/mediocregopher/radix/v4"

	"github.com/pilinux/gorest/example2/internal/database/model"
)

// KeyValueRepo provides methods for key-value operations.
type KeyValueRepo struct {
	redisClient radix.Client
}

// NewKeyValueRepo returns a new KeyValueRepo instance.
func NewKeyValueRepo(redisClient radix.Client) *KeyValueRepo {
	return &KeyValueRepo{
		redisClient: redisClient,
	}
}

// KeyValueRepository defines the contract for key-value data operations.
type KeyValueRepository interface {
	KeyExists(ctx context.Context, key string) (bool, error)

	// key-value operations
	SetKeyValue(ctx context.Context, key string, value string) (string, error)
	GetKeyValue(ctx context.Context, key string) (string, error)
	DeleteKeyValue(ctx context.Context, key string) (bool, error)

	// hash operations
	SetHashKeyValue(ctx context.Context, key string, value interface{}) error
	GetHashKeyValue(ctx context.Context, key string) (interface{}, error)
	DeleteHashKeyValue(ctx context.Context, key string, value interface{}) (bool, error)
}

// Compile-time check:
var _ KeyValueRepository = (*KeyValueRepo)(nil)

// KeyExists checks if a key exists in Redis.
func (r *KeyValueRepo) KeyExists(ctx context.Context, key string) (bool, error) {
	exists := 0
	if err := r.redisClient.Do(ctx, radix.FlatCmd(&exists, "EXISTS", key)); err != nil {
		return false, err
	}
	return exists > 0, nil
}

// SetKeyValue sets a key-value pair in Redis.
func (r *KeyValueRepo) SetKeyValue(ctx context.Context, key string, value string) (string, error) {
	result := ""
	if err := r.redisClient.Do(ctx, radix.FlatCmd(&result, "SET", key, value)); err != nil {
		return "", err
	}
	return result, nil
}

// GetKeyValue retrieves the value for a given key from Redis.
func (r *KeyValueRepo) GetKeyValue(ctx context.Context, key string) (string, error) {
	value := ""
	if err := r.redisClient.Do(ctx, radix.FlatCmd(&value, "GET", key)); err != nil {
		return "", err
	}
	return value, nil
}

// DeleteKeyValue deletes a key-value pair from Redis.
func (r *KeyValueRepo) DeleteKeyValue(ctx context.Context, key string) (bool, error) {
	ok := 0
	if err := r.redisClient.Do(ctx, radix.FlatCmd(&ok, "DEL", key)); err != nil {
		return false, err
	}
	return ok > 0, nil
}

// SetHashKeyValue sets a hash key-value pair in Redis.
// It uses the HSET command to set a field in a hash stored at key.
func (r *KeyValueRepo) SetHashKeyValue(ctx context.Context, key string, value interface{}) error {
	return r.redisClient.Do(ctx, radix.FlatCmd(nil, "HSET", key, value))
}

// GetHashKeyValue retrieves a value from a hash stored at key in Redis.
func (r *KeyValueRepo) GetHashKeyValue(ctx context.Context, key string) (interface{}, error) {
	var value model.KVHashValue
	if err := r.redisClient.Do(ctx, radix.FlatCmd(&value, "HGETALL", key)); err != nil {
		return nil, err
	}
	if reflect.DeepEqual(value, model.KVHashValue{}) {
		return nil, nil
	}
	return &value, nil
}

// DeleteHashKeyValue deletes a hash key-value pair from Redis.
func (r *KeyValueRepo) DeleteHashKeyValue(ctx context.Context, key string, value interface{}) (bool, error) {
	ok := 0
	if err := r.redisClient.Do(ctx, radix.FlatCmd(&ok, "HDEL", key, value)); err != nil {
		return false, err
	}
	return ok > 0, nil
}
