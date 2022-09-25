package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"
)

// GetHobbies - GET /hobbies
func GetHobbies(c *gin.Context) {
	resp, statusCode := handler.GetHobbies()

	renderer.Render(c, resp, statusCode)
}
