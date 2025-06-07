package handler

import "github.com/gin-gonic/gin"

// retrieve "authID" from the request
func getAuthID(c *gin.Context) uint64 {
	return c.GetUint64("authID")
}
