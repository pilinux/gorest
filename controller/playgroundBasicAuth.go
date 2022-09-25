package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/lib/renderer"
)

// AccessResource - can be accessed by basic auth
func AccessResource(c *gin.Context) {
	renderer.Render(c, gin.H{"msg": "access granted!"}, http.StatusOK)
}
