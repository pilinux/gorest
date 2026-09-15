// Package handler contains the thin HTTP layer of example3: it binds requests,
// calls the service layer and renders the response. It holds no business logic.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	grenderer "github.com/pilinux/gorest/lib/renderer"
)

// APIStatus reports that the API is live.
//
// Endpoint: GET / and GET /health
func APIStatus(c *gin.Context) {
	grenderer.Render(c, gin.H{"message": "live"}, http.StatusOK)
}
