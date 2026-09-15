package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	grenderer "github.com/pilinux/gorest/lib/renderer"

	"github.com/pilinux/gorest/example3/internal/database/model"
	"github.com/pilinux/gorest/example3/internal/service"
)

// TextCryptAPI exposes the text and number crypto endpoints.
type TextCryptAPI struct {
	svc *service.TextCryptService
}

// NewTextCryptAPI returns a new TextCryptAPI.
func NewTextCryptAPI(svc *service.TextCryptService) *TextCryptAPI {
	return &TextCryptAPI{svc: svc}
}

// EncryptText encrypts a plaintext string.
//
// Endpoint: POST /api/v1/crypto/text/encrypt
//
// Body: {"plaintext": "..."}
func (api *TextCryptAPI) EncryptText(c *gin.Context) {
	var req model.TextEncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := api.svc.EncryptText(req.Plaintext)
	grenderer.Render(c, resp, statusCode)
}

// DecryptText decrypts a token produced by EncryptText.
//
// Endpoint: POST /api/v1/crypto/text/decrypt
//
// Body: {"ciphertext": "..."}
func (api *TextCryptAPI) DecryptText(c *gin.Context) {
	var req model.DecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := api.svc.DecryptText(req.Ciphertext)
	grenderer.Render(c, resp, statusCode)
}

// EncryptNumber encrypts a signed integer.
//
// Endpoint: POST /api/v1/crypto/number/encrypt
//
// Body: {"number": 123}
func (api *TextCryptAPI) EncryptNumber(c *gin.Context) {
	var req model.NumberEncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := api.svc.EncryptNumber(req.Number)
	grenderer.Render(c, resp, statusCode)
}

// DecryptNumber decrypts a token produced by EncryptNumber.
//
// Endpoint: POST /api/v1/crypto/number/decrypt
//
// Body: {"ciphertext": "..."}
func (api *TextCryptAPI) DecryptNumber(c *gin.Context) {
	var req model.DecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := api.svc.DecryptNumber(req.Ciphertext)
	grenderer.Render(c, resp, statusCode)
}
