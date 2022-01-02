package controller

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/structs"
)

// render - render response
func render(c *gin.Context, data interface{}, statusCode int, htmlTpl ...string) {
	if len(htmlTpl) > 0 {
		reqType := c.Request.Header.Get("Accept")

		// Server HTML
		if strings.Contains(reqType, "text/html") {
			c.Set("template", htmlTpl[0])
			c.Set("data", structs.Map(data))
			return
		}
	}

	if statusCode >= 400 {
		c.AbortWithStatusJSON(statusCode, data)
		return
	}

	// Respond with JSON
	c.JSON(statusCode, data)
}
