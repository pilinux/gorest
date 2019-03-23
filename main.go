package main

import (
	//"fmt"

	"github.com/piLinux/GoREST/config"
	"github.com/piLinux/GoREST/controller"
	"github.com/piLinux/GoREST/database"
	"github.com/piLinux/GoREST/lib/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	configure := config.ConfigMain()
	db := database.InitDB()
	defer db.Close()

	// Debugging - environment variables
	/*
		fmt.Println(configure.Server.ServerPort)
		fmt.Println(configure.Database.DbDriver)
		fmt.Println(configure.Database.DbUser)
		fmt.Println(configure.Database.DbPass)
		fmt.Println(configure.Database.DbName)
		fmt.Println(configure.Database.DbHost)
		fmt.Println(configure.Database.DbPort)
	*/

	gin.SetMode(gin.ReleaseMode) // Omit this line to enable debug mode
	router := gin.Default()
	router.Use(middleware.CORS())

	// API:v1.0
	// Non-protected routes
	v1 := router.Group("/api/v1/")
	{
		// User
		v1.GET("users", controller.GetUsers)
		v1.GET("users/:id", controller.GetUser)
		v1.POST("users", controller.CreateUser)
		v1.PUT("users/:id", controller.UpdateUser)
		v1.DELETE("users/:id", controller.DeleteUser)

		// Post
		v1.GET("posts", controller.GetPosts)
		v1.GET("posts/:id", controller.GetPost)
		v1.POST("posts", controller.CreatePost)
		v1.PUT("posts/:id", controller.UpdatePost)
		v1.DELETE("posts/:id", controller.DeletePost)
	}

	router.Run(":" + configure.Server.ServerPort)

}
