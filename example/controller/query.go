package controller

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// QueryString - basic implementation
func QueryString(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query != "" {
		c.JSON(200, gin.H{"msg": query})
		return
	}

	c.JSON(200, gin.H{"msg": "query string is missing"})
}
