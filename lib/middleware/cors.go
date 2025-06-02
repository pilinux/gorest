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
// Copyright (c) 2022-2025 pilinux

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	rs "github.com/rs/cors"
	cors "github.com/rs/cors/wrapper/gin"
)

// CORSPolicy struct to handle all policies
type CORSPolicy struct {
	Key   string
	Value string
}

// CORS - Cross-Origin Resource Sharing
func CORS(cp []CORSPolicy) gin.HandlerFunc {
	// convert CORSPolicy slice to rs/cors.Options
	var allowedOrigins []string
	var allowedMethods []string
	var allowedHeaders []string
	var exposedHeaders []string
	var maxAge int
	var maxAgeSet bool
	var allowCredentials bool

	// helper function to deduplicate string slices
	deduplicate := func(input []string) []string {
		seen := make(map[string]struct{})
		result := make([]string, 0, len(input))
		for _, v := range input {
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				result = append(result, v)
			}
		}
		return result
	}

	for _, policy := range cp {
		key := strings.ToLower(policy.Key)
		switch key {
		case "access-control-allow-origin":
			for _, o := range strings.Split(policy.Value, ",") {
				trimmed := strings.TrimSpace(o)
				if trimmed != "" {
					allowedOrigins = append(allowedOrigins, trimmed)
				}
			}
		case "access-control-allow-methods":
			for _, m := range strings.Split(policy.Value, ",") {
				trimmed := strings.TrimSpace(m)
				if trimmed != "" {
					allowedMethods = append(allowedMethods, trimmed)
				}
			}
		case "access-control-allow-headers":
			for _, h := range strings.Split(policy.Value, ",") {
				trimmed := strings.TrimSpace(h)
				if trimmed != "" {
					allowedHeaders = append(allowedHeaders, trimmed)
				}
			}
		case "access-control-expose-headers":
			for _, h := range strings.Split(policy.Value, ",") {
				trimmed := strings.TrimSpace(h)
				if trimmed != "" {
					exposedHeaders = append(exposedHeaders, trimmed)
				}
			}
		case "access-control-max-age":
			if v, err := strconv.Atoi(policy.Value); err == nil {
				maxAge = v
				maxAgeSet = true
			}
		case "access-control-allow-credentials":
			if strings.ToLower(policy.Value) == "true" {
				allowCredentials = true
			}
		}
	}

	// deduplicate allowed lists
	allowedOrigins = deduplicate(allowedOrigins)
	allowedMethods = deduplicate(allowedMethods)
	allowedHeaders = deduplicate(allowedHeaders)
	exposedHeaders = deduplicate(exposedHeaders)

	// set default maxAge if not set by policy
	if !maxAgeSet {
		maxAge = 600 // default to 10 minutes
	}

	// prevent insecure CORS config: credentials + wildcard origin
	if allowCredentials {
		if len(allowedOrigins) == 0 {
			// if no origins are specified, by default all origins are allowed
			// which is not allowed with credentials
			return func(c *gin.Context) {
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					"CORS misconfiguration: CORS_CREDENTIALS=true with CORS_ORIGIN=* is forbidden by the CORS spec",
				)
			}
		}

		if strings.Contains(strings.Join(allowedOrigins, ", "), "*") {
			// if any origin is "*", return error
			// this is a security risk as it allows any origin to access the resource with credentials
			// according to the CORS spec, this is forbidden
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#access-control-allow-credentials
			return func(c *gin.Context) {
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					"CORS misconfiguration: CORS_CREDENTIALS=true with CORS_ORIGIN=* is forbidden by the CORS spec",
				)
			}
		}
	}

	// if no allowed methods, set to default
	if len(allowedMethods) == 0 {
		allowedMethods = []string{"OPTIONS"}
	}
	// if no allowed headers, set to default
	if len(allowedHeaders) == 0 {
		allowedHeaders = []string{"Content-Type"}
	}

	options := rs.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   allowedMethods,
		AllowedHeaders:   allowedHeaders,
		ExposedHeaders:   exposedHeaders,
		MaxAge:           maxAge,
		AllowCredentials: allowCredentials,
	}

	corsHandler := cors.New(options)

	return func(c *gin.Context) {
		for _, _cp := range cp {
			key := strings.ToLower(_cp.Key)
			switch key {
			case "x-content-type-options":
				c.Writer.Header().Set(_cp.Key, _cp.Value)
			case "x-frame-options":
				c.Writer.Header().Set(_cp.Key, _cp.Value)
			case "referrer-policy":
				c.Writer.Header().Set(_cp.Key, _cp.Value)
			case "content-security-policy":
				c.Writer.Header().Set(_cp.Key, _cp.Value)
			case "timing-allow-origin":
				origin := strings.TrimSpace(c.Request.Header.Get("Origin"))
				if strings.Contains(_cp.Value, "*") {
					c.Writer.Header().Set(_cp.Key, origin)
				} else {
					allowed := false
					for _, o := range strings.Split(_cp.Value, ",") {
						if strings.TrimSpace(o) == origin {
							allowed = true
							break
						}
					}
					if allowed {
						c.Writer.Header().Set(_cp.Key, origin)
					}
				}
			case "strict-transport-security":
				c.Writer.Header().Set(_cp.Key, _cp.Value)
			}
		}

		// required for browser-based HTTP clients
		if c.Request.Method == "OPTIONS" {
			// set Access-Control-Allow-Methods
			c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))

			// set Access-Control-Allow-Headers if present
			if len(allowedHeaders) > 0 {
				c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
			}

			// set Access-Control-Max-Age if set by policy
			if maxAgeSet {
				c.Writer.Header().Set("Access-Control-Max-Age", strconv.Itoa(maxAge))
			}

			corsHandler(c)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		corsHandler(c)
	}
}
