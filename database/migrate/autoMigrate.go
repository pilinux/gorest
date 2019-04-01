// autoMigrate.go needs to be executed only when it is required

package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/piLinux/GoREST/database"
	"github.com/piLinux/GoREST/database/model"
)

// Load all the models
type user model.User
type post model.Post
type userPost model.UserPost

var db *gorm.DB
var errorState int

func main() {
	/*
	** 0 = default/no error
	** 1 = error
	**/
	errorState = 0

	db = database.InitDB()
	defer db.Close()

	// Auto migration
	/*
		- Automatically migrate schema
		- Only create tables with missing columns and missing indexes
		- Will not change/delete any existing columns and their types
	*/

	// Careful! It will drop all the tables!
	dropAllTables()

	// Automatically migrate all the tables
	migrateTables()

	// Manually set foreign keys for MySQL DB
	setPkFk()

	if errorState == 0 {
		fmt.Println("Auto migration is completed!")
	} else {
		fmt.Println("Auto migration failed!")
	}
}

func dropAllTables() {
	// Careful! It will drop all the tables!
	if err := db.DropTableIfExists(&userPost{}, &user{}, &post{}).Error; err != nil {
		errorState = 1
		fmt.Println(err)
	} else {
		fmt.Println("Old tables are deleted!")
	}
}

func migrateTables() {
	// db.Set() --> add table suffix during auto migration
	if err := db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&user{},
		&post{}, &userPost{}).Error; err != nil {
		errorState = 1
		fmt.Println(err)
	} else {
		fmt.Println("New tables are  migrated successfully!")
	}
}

func setPkFk() {
	// Manually set foreign key for MySQL DB
	if err := db.Model(&userPost{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE").Error; err != nil {
		errorState = 1
		fmt.Println(err)
	}

	if err := db.Model(&userPost{}).AddForeignKey("post_id", "posts(id)", "CASCADE", "CASCADE").Error; err != nil {
		errorState = 1
		fmt.Println(err)
	}
}
