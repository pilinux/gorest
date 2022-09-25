// main function of the application
package main

import (
	"fmt"

	gconfig "github.com/pilinux/gorest/config"
	gdatabase "github.com/pilinux/gorest/database"

	"github.com/pilinux/gorest/example/database/migrate"
	"github.com/pilinux/gorest/example/router"
)

func main() {
	configure := gconfig.Config()

	if configure.Database.RDBMS.Activate == gconfig.Activated {
		// Initialize RDBMS client
		if err := gdatabase.InitDB().Error; err != nil {
			fmt.Println(err)
			return
		}

		// Drop all tables from DB
		/*
			if err := migrate.DropAllTables(); err != nil {
				fmt.Println(err)
				return
			}
		*/

		// Start DB migration
		if err := migrate.StartMigration(*configure); err != nil {
			fmt.Println(err)
			return
		}

		// Manually set foreign key for MySQL
		// execute it only when required!
		/*
			if err := migrate.SetPkFk(); err != nil {
				fmt.Println(err)
				return
			}
		*/
	}

	if configure.Database.REDIS.Activate == gconfig.Activated {
		// Initialize REDIS client
		if _, err := gdatabase.InitRedis(); err != nil {
			fmt.Println(err)
			return
		}
	}

	if configure.Database.MongoDB.Activate == gconfig.Activated {
		// Initialize MONGO client
		if _, err := gdatabase.InitMongo(); err != nil {
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
