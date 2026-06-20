// Package renderer uses a template engine to
// render and serve HTML pages.
package renderer

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/structs"
)

// Render renders a response in JSON format
// or uses a templating engine to serve HTML pages.
func Render(c *gin.Context, data any, statusCode int, htmlTpl ...string) {
	if len(htmlTpl) > 0 {
		reqType := c.Request.Header.Get("Accept")

		// Server HTML
		if strings.Contains(reqType, "text/html") {
			// apply the status code before handing off to the template
			// middleware so error pages are not served as HTTP 200
			c.Status(statusCode)
			c.Set("template", htmlTpl[0])
			// structs.Map panics on non-struct data; pass such values through
			model := data
			if structs.IsStruct(data) {
				model = structs.Map(data)
			}
			c.Set("data", model)
			return
		}
	}

	if statusCode >= 400 {
		c.AbortWithStatusJSON(statusCode, data)
		return
	}

	// Respond with JSON
	c.SecureJSON(statusCode, data)
}
