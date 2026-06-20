// Package renderer uses a template engine to
// render and serve HTML pages.
package renderer

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 - 2026 pilinux

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/structs"
)

// Render renders a response in JSON format
// or uses a templating engine to serve HTML pages.
func Render(c *gin.Context, data any, statusCode int, htmlTpl ...string) {
	if len(htmlTpl) > 0 {
		accept := c.Request.Header.Get("Accept")

		// Serve HTML only when the client actually prefers it over JSON,
		// honouring the Accept header's media ranges and q-values
		if acceptQuality(accept, "text", "html") > acceptQuality(accept, "application", "json") {
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

// acceptQuality returns the q-value the Accept header assigns to the media type
// "typ/sub". It parses comma-separated media ranges with their q-values and
// applies the most specific matching range (exact > "typ/*" > "*/*"). It returns
// 0 when the media type is not acceptable.
func acceptQuality(accept, typ, sub string) float64 {
	bestSpec := -1
	bestQ := 0.0

	for part := range strings.SplitSeq(accept, ",") {
		mediaType, params, _ := strings.Cut(part, ";")
		rTyp, rSub, ok := strings.Cut(strings.TrimSpace(mediaType), "/")
		if !ok {
			continue
		}
		rTyp = strings.TrimSpace(rTyp)
		rSub = strings.TrimSpace(rSub)

		// determine how specifically this range matches the wanted media type
		var spec int
		switch {
		case rTyp == typ && rSub == sub:
			spec = 2
		case rTyp == typ && rSub == "*":
			spec = 1
		case rTyp == "*" && rSub == "*":
			spec = 0
		default:
			continue
		}

		// q-value defaults to 1.0 when not specified
		q := 1.0
		for p := range strings.SplitSeq(params, ";") {
			p = strings.TrimSpace(p)
			if v, ok := strings.CutPrefix(p, "q="); ok {
				if parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
					q = parsed
				}
			}
		}

		if spec > bestSpec {
			bestSpec = spec
			bestQ = q
		}
	}

	return bestQ
}
