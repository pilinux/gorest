package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/example3/internal/handler"
	"github.com/pilinux/gorest/example3/internal/service"
)

// newTextEngine builds a gin engine with the text and number crypto routes.
func newTextEngine(t *testing.T) *gin.Engine {
	t.Helper()
	api := handler.NewTextCryptAPI(service.NewTextCryptService(newLoadedKeyManager(t)))

	r := gin.New()
	r.POST("/text/encrypt", api.EncryptText)
	r.POST("/text/decrypt", api.DecryptText)
	r.POST("/number/encrypt", api.EncryptNumber)
	r.POST("/number/decrypt", api.DecryptNumber)
	return r
}

// doJSON performs a JSON POST and returns the recorder.
func doJSON(t *testing.T, r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// messageString extracts message.<field> as a string from a JSON response body.
func messageField(t *testing.T, body []byte, field string) string {
	t.Helper()
	var parsed struct {
		Message map[string]any `json:"message"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("unmarshal error: %v (body: %s)", err, body)
	}
	v, _ := parsed.Message[field].(string)
	return v
}

func TestHandler_TextRoundTrip(t *testing.T) {
	r := newTextEngine(t)

	// encrypt
	w := doJSON(t, r, "/text/encrypt", `{"plaintext":"hello world"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("encrypt status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	token := messageField(t, w.Body.Bytes(), "ciphertext")
	if token == "" {
		t.Fatal("empty ciphertext token")
	}

	// decrypt
	body, _ := json.Marshal(map[string]string{"ciphertext": token})
	w = doJSON(t, r, "/text/decrypt", string(body))
	if w.Code != http.StatusOK {
		t.Fatalf("decrypt status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	if got := messageField(t, w.Body.Bytes(), "plaintext"); got != "hello world" {
		t.Errorf("plaintext = %q, want \"hello world\"", got)
	}
}

func TestHandler_TextEncrypt_BadRequest(t *testing.T) {
	r := newTextEngine(t)

	// malformed JSON
	if w := doJSON(t, r, "/text/encrypt", `{`); w.Code != http.StatusBadRequest {
		t.Errorf("malformed JSON status = %d, want 400", w.Code)
	}
	// empty plaintext
	if w := doJSON(t, r, "/text/encrypt", `{"plaintext":""}`); w.Code != http.StatusBadRequest {
		t.Errorf("empty plaintext status = %d, want 400", w.Code)
	}
}

func TestHandler_TextNumber_BodyTooLarge(t *testing.T) {
	r := newTextEngine(t)

	// one byte over the 8 MiB body cap, as the only field value
	filler := strings.Repeat("a", 8<<20)
	tests := []struct {
		name string
		path string
		body string
	}{
		{name: "text encrypt", path: "/text/encrypt", body: `{"plaintext":"` + filler + `"}`},
		{name: "text decrypt", path: "/text/decrypt", body: `{"ciphertext":"` + filler + `"}`},
		{name: "number encrypt", path: "/number/encrypt", body: `{"number":1,"pad":"` + filler + `"}`},
		{name: "number decrypt", path: "/number/decrypt", body: `{"ciphertext":"` + filler + `"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if w := doJSON(t, r, tt.path, tt.body); w.Code != http.StatusRequestEntityTooLarge {
				t.Errorf("status = %d, want 413 (body: %s)", w.Code, w.Body.String())
			}
		})
	}
}

func TestHandler_TextDecrypt_BadToken(t *testing.T) {
	r := newTextEngine(t)
	if w := doJSON(t, r, "/text/decrypt", `{"ciphertext":"not-valid"}`); w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandler_NumberRoundTrip(t *testing.T) {
	r := newTextEngine(t)

	w := doJSON(t, r, "/number/encrypt", `{"number":123456789}`)
	if w.Code != http.StatusOK {
		t.Fatalf("encrypt status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	token := messageField(t, w.Body.Bytes(), "ciphertext")

	body, _ := json.Marshal(map[string]string{"ciphertext": token})
	w = doJSON(t, r, "/number/decrypt", string(body))
	if w.Code != http.StatusOK {
		t.Fatalf("decrypt status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}

	// number decodes as a JSON number (float64 in a generic map)
	var parsed struct {
		Message struct {
			Number float64 `json:"number"`
		} `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if parsed.Message.Number != 123456789 {
		t.Errorf("number = %v, want 123456789", parsed.Message.Number)
	}
}

func TestHandler_NumberEncrypt_Missing(t *testing.T) {
	r := newTextEngine(t)
	// no "number" field -> nil pointer -> 400
	if w := doJSON(t, r, "/number/encrypt", `{}`); w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
