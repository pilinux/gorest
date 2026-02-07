package model

// HTTPResponse represents the final response sent to API consumers.
type HTTPResponse struct {
	Message any `json:"message,omitempty"`
}
