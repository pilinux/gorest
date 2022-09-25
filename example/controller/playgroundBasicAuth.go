package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	grenderer "github.com/pilinux/gorest/lib/renderer"
)

// AccessResource - can be accessed by basic auth
func AccessResource(c *gin.Context) {
	grenderer.Render(c, gin.H{"message": "access granted!"}, http.StatusOK)
}
