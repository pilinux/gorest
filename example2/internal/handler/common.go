package handler

import "github.com/gin-gonic/gin"

// getAuthID retrieves the authID from the request.
func getAuthID(c *gin.Context) uint64 {
	return c.GetUint64("authID")
}
