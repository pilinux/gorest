package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	grenderer "github.com/pilinux/gorest/lib/renderer"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/example3/internal/database/model"
	"github.com/pilinux/gorest/example3/internal/service"
)

// maxJSONBodySize limits a text or number request body before it is decoded.
// The service accepts up to 1 MiB of text, and JSON escaping can turn one byte
// into six (\u00XX), so 8 MiB fits that with a little room to spare.
const maxJSONBodySize = 8 << 20 // 8 MiB

// TextCryptAPI exposes the text and number crypto endpoints.
type TextCryptAPI struct {
	svc *service.TextCryptService
}

// NewTextCryptAPI returns a new TextCryptAPI.
func NewTextCryptAPI(svc *service.TextCryptService) *TextCryptAPI {
	return &TextCryptAPI{svc: svc}
}

// bindJSON decodes a size-capped JSON body into req. On failure it renders the
// error response and returns false.
func bindJSON(c *gin.Context, req any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxJSONBodySize)
	if err := c.ShouldBindJSON(req); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			grenderer.Render(c, gin.H{"message": "request body is too large"}, http.StatusRequestEntityTooLarge)
			return false
		}
		// the binder names fields and Go types, so it is logged, not sent
		log.WithContext(c.Request.Context()).WithError(err).Error("bindJSON.h.1")
		grenderer.Render(c, gin.H{"message": "invalid request body"}, http.StatusBadRequest)
		return false
	}
	return true
}

// EncryptText encrypts a plaintext string.
//
// Endpoint: POST /api/v1/crypto/text/encrypt
//
// Body: {"plaintext": "..."}
func (api *TextCryptAPI) EncryptText(c *gin.Context) {
	var req model.TextEncryptRequest
	if !bindJSON(c, &req) {
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
	if !bindJSON(c, &req) {
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
	if !bindJSON(c, &req) {
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
	if !bindJSON(c, &req) {
		return
	}

	resp, statusCode := api.svc.DecryptNumber(req.Ciphertext)
	grenderer.Render(c, resp, statusCode)
}
