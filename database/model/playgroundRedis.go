package model

// RedisData - key:value
type RedisData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RedisDataHash - key:hashValues
type RedisDataHash struct {
	Key   string `json:"key"`
	Value RedisDataHashValue
}

// RedisDataHashValue - values
type RedisDataHashValue struct {
	Value1 string `json:"value1"`
	Value2 string `json:"value2"`
	Value3 string `json:"value3"`
	Value4 string `json:"value4"`
}
