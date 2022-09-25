// Package migrate to migrate the schema
package migrate

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database/model"
)

// Load all the models
type auth model.Auth
type twoFA model.TwoFA

var db *gorm.DB

// DropAllTables - careful! It will drop all the tables!
func DropAllTables() error {
	if err := db.Migrator().DropTable(
		&twoFA{},
		&auth{},
	); err != nil {
		return err
	}

	fmt.Println("old tables are deleted!")
	return nil
}

// StartMigration - automatically migrate all the tables
// - Only create tables with missing columns and missing indexes
// - Will not change/delete any existing columns and their types
func StartMigration() error {
	configureDB := config.GetConfig().Database.RDBMS
	driver := configureDB.Env.Driver

	if driver == "mysql" {
		// db.Set() --> add table suffix during auto migration
		if err := db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
			&auth{},
			&twoFA{},
		); err != nil {
			return err
		}

		fmt.Println("new tables are  migrated successfully!")
		return nil
	}

	if err := db.AutoMigrate(
		&auth{},
		&twoFA{},
	); err != nil {
		return err
	}

	fmt.Println("New tables are  migrated successfully!")
	return nil
}
