package model

// Request and response bodies for the text and number endpoints.

// TextEncryptRequest is the body of the text-encrypt endpoint.
type TextEncryptRequest struct {
	Plaintext string `json:"plaintext"`
}

// NumberEncryptRequest is the body of the number-encrypt endpoint. Number is a
// pointer, so a missing number is not mistaken for 0.
type NumberEncryptRequest struct {
	Number *int64 `json:"number"`
}

// DecryptRequest is the body of both decrypt endpoints: the base64 token an
// encrypt call returned.
type DecryptRequest struct {
	Ciphertext string `json:"ciphertext"`
}

// CiphertextResponse is returned by both encrypt endpoints.
type CiphertextResponse struct {
	Ciphertext string `json:"ciphertext"`
}

// PlaintextResponse is returned by the text-decrypt endpoint.
type PlaintextResponse struct {
	Plaintext string `json:"plaintext"`
}

// NumberResponse is returned by the number-decrypt endpoint.
type NumberResponse struct {
	Number int64 `json:"number"`
}
