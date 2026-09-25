package service

import (
	"net/http"

	gmodel "github.com/pilinux/gorest/database/model"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/example3/internal/database/model"
)

// maxTextLen is the largest plaintext the text endpoint accepts, so this demo
// API is not used to encrypt huge blobs.
const maxTextLen = 1 << 20 // 1 MiB

// masterKeyProvider is the part of KeyManager the crypto services need. Using
// an interface keeps them easy to test.
type masterKeyProvider interface {
	MasterKey() ([]byte, error)
}

// TextCryptService encrypts and decrypts text and numbers using the master key.
type TextCryptService struct {
	keys masterKeyProvider
}

// NewTextCryptService returns a new TextCryptService.
func NewTextCryptService(keys masterKeyProvider) *TextCryptService {
	return &TextCryptService{keys: keys}
}

// EncryptText encrypts a string and returns a base64 token.
func (s *TextCryptService) EncryptText(plaintext string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	if plaintext == "" {
		httpResponse.Message = "plaintext is required"
		httpStatusCode = http.StatusBadRequest
		return
	}
	if len(plaintext) > maxTextLen {
		httpResponse.Message = "plaintext is too large"
		httpStatusCode = http.StatusRequestEntityTooLarge
		return
	}

	masterKey, err := s.keys.MasterKey()
	if err != nil {
		log.WithError(err).Error("EncryptText.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	token, err := scheme.SealStringAAD(masterKey, plaintext, []byte(textAADLabel))
	if err != nil {
		log.WithError(err).Error("EncryptText.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = model.CiphertextResponse{Ciphertext: token}
	httpStatusCode = http.StatusOK
	return
}

// DecryptText turns a base64 token back into the original string.
func (s *TextCryptService) DecryptText(token string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	if token == "" {
		httpResponse.Message = "ciphertext is required"
		httpStatusCode = http.StatusBadRequest
		return
	}

	masterKey, err := s.keys.MasterKey()
	if err != nil {
		log.WithError(err).Error("DecryptText.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	plaintext, err := scheme.OpenStringAAD(masterKey, token, []byte(textAADLabel))
	if err != nil {
		// a bad or tampered token is the client's error, not the server's
		httpResponse.Message = "unable to decrypt: invalid or corrupted ciphertext"
		httpStatusCode = http.StatusBadRequest
		return
	}

	httpResponse.Message = model.PlaintextResponse{Plaintext: plaintext}
	httpStatusCode = http.StatusOK
	return
}

// EncryptNumber encrypts an integer and returns a base64 token. number is a
// pointer, so a missing value is rejected instead of becoming 0.
func (s *TextCryptService) EncryptNumber(number *int64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	if number == nil {
		httpResponse.Message = "number is required"
		httpStatusCode = http.StatusBadRequest
		return
	}

	masterKey, err := s.keys.MasterKey()
	if err != nil {
		log.WithError(err).Error("EncryptNumber.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	token, err := scheme.SealInt64AAD(masterKey, *number, []byte(int64AADLabel))
	if err != nil {
		log.WithError(err).Error("EncryptNumber.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = model.CiphertextResponse{Ciphertext: token}
	httpStatusCode = http.StatusOK
	return
}

// DecryptNumber turns a base64 token back into the original integer.
func (s *TextCryptService) DecryptNumber(token string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	if token == "" {
		httpResponse.Message = "ciphertext is required"
		httpStatusCode = http.StatusBadRequest
		return
	}

	masterKey, err := s.keys.MasterKey()
	if err != nil {
		log.WithError(err).Error("DecryptNumber.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	number, err := scheme.OpenInt64AAD(masterKey, token, []byte(int64AADLabel))
	if err != nil {
		httpResponse.Message = "unable to decrypt: invalid or corrupted ciphertext"
		httpStatusCode = http.StatusBadRequest
		return
	}

	httpResponse.Message = model.NumberResponse{Number: number}
	httpStatusCode = http.StatusOK
	return
}
