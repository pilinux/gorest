package model

// HTTPResponse - final response to the api consumers
type HTTPResponse struct {
	Message interface{} `json:"message,omitempty"`
}
