// main function of the application
package main

import (
	"fmt"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/router"
)

func main() {
	configure := config.Config()

	if configure.Database.RDBMS.Activate == config.Activated {
		// Initialize RDBMS client
		if err := database.InitDB().Error; err != nil {
			fmt.Println(err)
			return
		}
	}

	if configure.Database.REDIS.Activate == config.Activated {
		// Initialize REDIS client
		if _, err := database.InitRedis(); err != nil {
			fmt.Println(err)
			return
		}
	}

	if configure.Database.MongoDB.Activate == config.Activated {
		// Initialize MONGO client
		if _, err := database.InitMongo(); err != nil {
			fmt.Println(err)
			return
		}
	}

	r, err := router.SetupRouter(configure)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Attaches the router to a http.Server and starts listening and serving HTTP requests
	err = r.Run(":" + configure.Server.ServerPort)
	if err != nil {
		fmt.Println(err)
		return
	}
}
