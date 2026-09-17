package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/example3/internal/handler"
)

// TestAPIStatus - the health endpoint replies 200 with a "live" body.
func TestAPIStatus(t *testing.T) {
	r := gin.New()
	r.GET("/health", handler.APIStatus)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "live") {
		t.Errorf("body = %q, want it to contain \"live\"", w.Body.String())
	}
}
