// autoMigrate.go needs to be executed only when it is required

package main

import (
	"fmt"

	"github.com/piLinux/GoREST/database"
	"github.com/piLinux/GoREST/database/model"

	"github.com/jinzhu/gorm"
)

var db *gorm.DB

func main() {
	db := database.InitDB()
	defer db.Close()

	// Auto migration
	/*
		- Automatically migrate schema
		- Only create tables with missing columns and missing indexes
		- Will not change/delete any existing columns and their types
	*/

	// Load all the models
	type User = model.User
	type Post = model.Post
	type UserPost = model.UserPost

	// Craeful! It will drop all tables!
	if err := db.DropTableIfExists(&UserPost{}, &User{}, &Post{}).Error; err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Old tables are deleted!")
	}

	// db.Set() --> add table suffix during auto migration
	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&User{}, &Post{}, &UserPost{})

	// Manually set foreign key for MySQL DB
	db.Model(&UserPost{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	db.Model(&UserPost{}).AddForeignKey("post_id", "posts(id)", "CASCADE", "CASCADE")

	fmt.Println("Auto migration is completed!")
}
