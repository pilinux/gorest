package database

import (
	"fmt"
	"github.com/GoREST/config"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var DB *gorm.DB
var err error

func InitDB() *gorm.DB {
	var db = DB

	configureDB := config.ConfigMain()

	driver := configureDB.Database.DbDriver
	username := configureDB.Database.DbUser
	password := configureDB.Database.DbPass
	database := configureDB.Database.DbName
	host := configureDB.Database.DbHost
	port := configureDB.Database.DbPort

	if driver == "mysql" {
		db, err = gorm.Open(driver, username+":"+password+"@tcp("+host+":"+port+")/"+database+"?charset=utf8&parseTime=True&loc=Local")
		if err != nil {
			fmt.Println("DB err: ", err)
		}
		// Only for debugging
		if err == nil {
			fmt.Println("DB connection successful!")
		}

	}

	DB = db

	return DB
}

// GetDB - get a connection
func GetDB() *gorm.DB {
	return DB
}
