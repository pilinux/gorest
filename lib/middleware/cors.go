// Package middleware contains:
//
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

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSPolicy struct to handle all policies
type CORSPolicy struct {
	Key   string
	Value string
}

// CORS - Cross-Origin Resource Sharing
func CORS(cp []CORSPolicy) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, _cp := range cp {
			c.Writer.Header().Set(_cp.Key, _cp.Value)
		}

		// required for browser-based HTTP clients
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
