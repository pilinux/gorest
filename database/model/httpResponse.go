package model

// HTTPResponse - final response to the api consumers
type HTTPResponse struct {
	Result interface{} `json:"result,omitempty"`
}
