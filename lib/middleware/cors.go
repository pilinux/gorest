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

// CORSConfig holds all CORS configuration values
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	MaxAge           int
	AllowCredentials bool
}

var (
	corsConfig CORSConfig
	maxAgeSet  bool
)

// CORS - Cross-Origin Resource Sharing
func CORS(cp []CORSPolicy) gin.HandlerFunc {
	ResetCORS()

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

	// convert CORSPolicy slice to rs/cors.Options
	for _, policy := range cp {
		key := strings.ToLower(policy.Key)
		switch key {
		case "access-control-allow-origin":
			for _, o := range strings.Split(policy.Value, ",") {
				trimmed := strings.TrimSpace(o)
				if trimmed != "" {
					corsConfig.AllowedOrigins = append(corsConfig.AllowedOrigins, trimmed)
				}
			}
		case "access-control-allow-methods":
			for _, m := range strings.Split(policy.Value, ",") {
				trimmed := strings.TrimSpace(m)
				if trimmed != "" {
					corsConfig.AllowedMethods = append(corsConfig.AllowedMethods, trimmed)
				}
			}
		case "access-control-allow-headers":
			for _, h := range strings.Split(policy.Value, ",") {
				trimmed := strings.TrimSpace(h)
				if trimmed != "" {
					corsConfig.AllowedHeaders = append(corsConfig.AllowedHeaders, trimmed)
				}
			}
		case "access-control-expose-headers":
			for _, h := range strings.Split(policy.Value, ",") {
				trimmed := strings.TrimSpace(h)
				if trimmed != "" {
					corsConfig.ExposedHeaders = append(corsConfig.ExposedHeaders, trimmed)
				}
			}
		case "access-control-max-age":
			if v, err := strconv.Atoi(policy.Value); err == nil {
				corsConfig.MaxAge = v
				maxAgeSet = true
			}
		case "access-control-allow-credentials":
			if strings.ToLower(policy.Value) == "true" {
				corsConfig.AllowCredentials = true
			}
		}
	}

	// deduplicate allowed lists
	corsConfig.AllowedOrigins = deduplicate(corsConfig.AllowedOrigins)
	corsConfig.AllowedMethods = deduplicate(corsConfig.AllowedMethods)
	corsConfig.AllowedHeaders = deduplicate(corsConfig.AllowedHeaders)
	corsConfig.ExposedHeaders = deduplicate(corsConfig.ExposedHeaders)

	// set default maxAge if not set by policy
	if !maxAgeSet {
		corsConfig.MaxAge = 600 // default to 10 minutes
	}

	// if no allowed methods, set to default
	if len(corsConfig.AllowedMethods) == 0 {
		corsConfig.AllowedMethods = []string{"OPTIONS"}
	}
	// if no allowed headers, set to default
	if len(corsConfig.AllowedHeaders) == 0 {
		corsConfig.AllowedHeaders = []string{"Content-Type"}
	}

	// prevent insecure CORS config: credentials + wildcard origin
	if corsConfig.AllowCredentials {
		if len(corsConfig.AllowedOrigins) == 0 {
			// if no origins are specified, by default all origins are allowed
			// which is not allowed with credentials
			return func(c *gin.Context) {
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					"CORS misconfiguration: CORS_CREDENTIALS=true with CORS_ORIGIN=* is forbidden by the CORS spec",
				)
			}
		}

		// if any origin is "*", return error
		// this is a security risk as it allows any origin to access the resource with credentials
		// according to the CORS spec, this is forbidden
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#access-control-allow-credentials
		if strings.Contains(strings.Join(corsConfig.AllowedOrigins, ", "), "*") {
			return func(c *gin.Context) {
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					"CORS misconfiguration: CORS_CREDENTIALS=true with CORS_ORIGIN=* is forbidden by the CORS spec",
				)
			}
		}

		// if any method is "*", return error
		if strings.Contains(strings.Join(corsConfig.AllowedMethods, ", "), "*") {
			return func(c *gin.Context) {
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					"CORS misconfiguration: CORS_CREDENTIALS=true with CORS_METHODS=* is forbidden by the CORS spec",
				)
			}
		}

		// if any header is "*", return error
		if strings.Contains(strings.Join(corsConfig.AllowedHeaders, ", "), "*") {
			return func(c *gin.Context) {
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					"CORS misconfiguration: CORS_CREDENTIALS=true with CORS_HEADERS=* is forbidden by the CORS spec",
				)
			}
		}

		// if any exposed header is "*", return error
		if strings.Contains(strings.Join(corsConfig.ExposedHeaders, ", "), "*") {
			return func(c *gin.Context) {
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					"CORS misconfiguration: CORS_CREDENTIALS=true with CORS_EXPOSED_HEADERS=* is forbidden by the CORS spec",
				)
			}
		}
	}

	options := rs.Options{
		AllowedOrigins:   corsConfig.AllowedOrigins,
		AllowedMethods:   corsConfig.AllowedMethods,
		AllowedHeaders:   corsConfig.AllowedHeaders,
		ExposedHeaders:   corsConfig.ExposedHeaders,
		MaxAge:           corsConfig.MaxAge,
		AllowCredentials: corsConfig.AllowCredentials,
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
			c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowedMethods, ", "))

			// set Access-Control-Allow-Headers if present
			if len(corsConfig.AllowedHeaders) > 0 {
				c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowedHeaders, ", "))
			}

			// set Access-Control-Max-Age if set by policy
			if maxAgeSet {
				c.Writer.Header().Set("Access-Control-Max-Age", strconv.Itoa(corsConfig.MaxAge))
			}

			corsHandler(c)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		corsHandler(c)
	}
}

// GetCORS returns all CORS configuration values in a struct
func GetCORS() CORSConfig {
	return corsConfig
}

// ResetCORS resets the CORS configuration
func ResetCORS() {
	corsConfig = CORSConfig{}
	maxAgeSet = false
}
