package model

// KeyValue - general model to process
// key-value pair
type KeyValue struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
