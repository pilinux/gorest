package model

// RedisData represents a key-value pair for Redis.
type RedisData struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// RedisDataHash represents a key with hash values for Redis.
type RedisDataHash struct {
	Key   string             `json:"key,omitempty"`
	Value RedisDataHashValue `json:"value,omitempty"`
}

// RedisDataHashValue holds the hash field values for RedisDataHash.
type RedisDataHashValue struct {
	Value1 string `json:"value1,omitempty"`
	Value2 string `json:"value2,omitempty"`
	Value3 string `json:"value3,omitempty"`
	Value4 string `json:"value4,omitempty"`
}
