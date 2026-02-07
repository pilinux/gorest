package model

import (
	"errors"
	"strings"
)

// KV represents a key-value pair.
type KV struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// Trim trims the KV struct to remove leading and trailing spaces.
func (kv *KV) Trim() error {
	kv.Key = strings.TrimSpace(kv.Key)
	kv.Value = strings.TrimSpace(kv.Value)

	if kv.Key == "" {
		return errors.New("key is required")
	}
	if kv.Value == "" {
		return errors.New("value is required")
	}

	return nil
}

// KVHash represents a key with hash values.
type KVHash struct {
	Key   string      `json:"key,omitempty"`
	Value KVHashValue `json:"value,omitempty"`
}

// KVHashValue holds the values for a hash key.
type KVHashValue struct {
	Value1 string `json:"value1,omitempty"`
	Value2 string `json:"value2,omitempty"`
	Value3 string `json:"value3,omitempty"`
	Value4 string `json:"value4,omitempty"`
}

// Trim trims the KVHash struct to remove leading and trailing spaces.
func (kvh *KVHash) Trim() error {
	kvh.Key = strings.TrimSpace(kvh.Key)
	return kvh.Value.Trim()
}

// Trim trims the KVHashValue struct to remove leading and trailing spaces.
func (kvhVal *KVHashValue) Trim() error {
	if kvhVal == nil {
		return errors.New("kvhVal is nil")
	}

	kvhVal.Value1 = strings.TrimSpace(kvhVal.Value1)
	kvhVal.Value2 = strings.TrimSpace(kvhVal.Value2)
	kvhVal.Value3 = strings.TrimSpace(kvhVal.Value3)
	kvhVal.Value4 = strings.TrimSpace(kvhVal.Value4)

	if kvhVal.Value1 == "" && kvhVal.Value2 == "" && kvhVal.Value3 == "" && kvhVal.Value4 == "" {
		return errors.New("at least one value is required")
	}

	return nil
}
