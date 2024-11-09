package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	grenderer "github.com/pilinux/gorest/lib/renderer"
)

// AccessResource - can be accessed by basic auth
func AccessResource(c *gin.Context) {
	// print all claims in JWT
	for k, v := range c.Keys {
		fmt.Println("key:", k, "|", "value:", v)
	}

	grenderer.Render(c, gin.H{"message": "access granted!"}, http.StatusOK)
}
