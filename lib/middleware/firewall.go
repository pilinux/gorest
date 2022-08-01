package middleware

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Firewall - whitelist/blacklist IPs
func Firewall(listType string, ipList string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the real client IP
		clientIP := c.ClientIP()

		if !strings.Contains(ipList, "*") {
			if listType == "whitelist" {
				if !strings.Contains(ipList, clientIP) {
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
			}

			if listType == "blacklist" {
				if strings.Contains(ipList, clientIP) {
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
			}
		}

		if strings.Contains(ipList, "*") {
			if listType == "blacklist" {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}

		c.Next()
	}
}
