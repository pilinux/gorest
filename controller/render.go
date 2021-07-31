package controller

import "github.com/gin-gonic/gin"

// render - render response
func render(c *gin.Context, data interface{}, statusCode int) {
	if statusCode >= 400 {
		c.AbortWithStatusJSON(statusCode, data)
		return
	}

	// Respond with JSON
	c.JSON(statusCode, data)

	/*
		// Reference: to implement other formats
		switch c.Request.Header.Get("Accept") {
		case "application/json":
			// Respond with JSON
			c.JSON(statusCode, data)
		case "application/xml":
			// Respond with XML
			c.XML(statusCode, data)
		default:
			// Respond with ...
			c...
		}
	*/
}
