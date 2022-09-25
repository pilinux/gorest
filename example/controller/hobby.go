package controller

import (
	"github.com/gin-gonic/gin"

	grenderer "github.com/pilinux/gorest/lib/renderer"

	"github.com/pilinux/gorest/example/handler"
)

// GetHobbies - GET /hobbies
func GetHobbies(c *gin.Context) {
	resp, statusCode := handler.GetHobbies()

	grenderer.Render(c, resp, statusCode)
}
