package renderer_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/lib/renderer"
)

// testPayload is a simple struct used as response data in tests.
type testPayload struct {
	Message string `json:"message,omitempty"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

// TestRender_JSONSuccess tests that Render writes a SecureJSON response
// when no template is provided and statusCode < 400.
func TestRender_JSONSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	data := testPayload{Message: "ok"}
	renderer.Render(c, data, http.StatusOK)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var got testPayload
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if got.Message != "ok" {
		t.Errorf("expected message %q, got %q", "ok", got.Message)
	}
}

// TestRender_JSONError tests that Render calls AbortWithStatusJSON
// when statusCode >= 400 and no template is provided.
func TestRender_JSONError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	data := testPayload{Message: "bad request"}
	renderer.Render(c, data, http.StatusBadRequest)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !c.IsAborted() {
		t.Error("expected context to be aborted")
	}

	var got testPayload
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if got.Message != "bad request" {
		t.Errorf("expected message %q, got %q", "bad request", got.Message)
	}
}

// TestRender_HTMLTemplate tests that Render sets the template and data
// on the gin context when a template is provided and Accept contains text/html.
func TestRender_HTMLTemplate(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Accept", "text/html")

	data := testPayload{Message: "hello"}
	renderer.Render(c, data, http.StatusOK, "index.html")

	tpl, exists := c.Get("template")
	if !exists {
		t.Fatal("expected 'template' to be set in context")
	}
	if tpl != "index.html" {
		t.Errorf("expected template %q, got %q", "index.html", tpl)
	}

	d, exists := c.Get("data")
	if !exists {
		t.Fatal("expected 'data' to be set in context")
	}
	dataMap, ok := d.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map[string]interface{}, got %T", d)
	}
	if dataMap["Message"] != "hello" {
		t.Errorf("expected Message %q, got %q", "hello", dataMap["Message"])
	}

	// no body should be written since the template engine handles rendering
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body, got %q", w.Body.String())
	}
}

// TestRender_TemplateWithNonHTMLAccept tests that when a template is provided
// but the Accept header does not contain text/html, Render falls through to
// JSON rendering.
func TestRender_TemplateWithNonHTMLAccept(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Accept", "application/json")

	data := testPayload{Message: "json fallback"}
	renderer.Render(c, data, http.StatusOK, "index.html")

	// template and data should NOT be set
	if _, exists := c.Get("template"); exists {
		t.Error("expected 'template' NOT to be set in context")
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var got testPayload
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if got.Message != "json fallback" {
		t.Errorf("expected message %q, got %q", "json fallback", got.Message)
	}
}

// TestRender_TemplateWithNonHTMLAcceptError tests that when a template is
// provided, Accept is not text/html, and statusCode >= 400, Render calls
// AbortWithStatusJSON.
func TestRender_TemplateWithNonHTMLAcceptError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Accept", "application/json")

	data := testPayload{Message: "not found"}
	renderer.Render(c, data, http.StatusNotFound, "error.html")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
	if !c.IsAborted() {
		t.Error("expected context to be aborted")
	}
}
