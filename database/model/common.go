package model

// KeyValue represents a general-purpose
// key-value pair.
type KeyValue struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
