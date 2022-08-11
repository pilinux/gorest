// Package middleware contains:
// - CORS
// - Application firewall
// - Pongo2 template engine
// - JWT
// - Sentry logger
// - Two-factor auth validator
package middleware

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import "github.com/gin-gonic/gin"

// CORS - Cross-Origin Resource Sharing
// origin: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
// credentials: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials
// headers: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers
// methods: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods
// maxAge: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age
// General recommendation for dev environment:
// ----------------------------------
// origin: "*"
// credentials: "true"
// headers: "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"
// methods: "GET, POST, PUT, PATCH, DELETE"
// maxAge: 300
func CORS(origin, credentials, headers, methods, maxAge string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", credentials)
		c.Writer.Header().Set("Access-Control-Allow-Headers", headers)
		c.Writer.Header().Set("Access-Control-Allow-Methods", methods)
		c.Writer.Header().Set("Access-Control-Max-Age", maxAge)
		c.Next()
	}
}
