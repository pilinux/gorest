// Package migrate handles database schema migration for the example2 application.
package migrate

import (
	"fmt"

	gconfig "github.com/pilinux/gorest/config"
	gdb "github.com/pilinux/gorest/database"
	gmodel "github.com/pilinux/gorest/database/model"

	"github.com/pilinux/gorest/example2/internal/database/model"
)

// load all the models.
type auth gmodel.Auth
type twoFA gmodel.TwoFA
type twoFABackup gmodel.TwoFABackup
type tempEmail gmodel.TempEmail
type user model.User
type post model.Post
type hobby model.Hobby
type userHobby model.UserHobby

// DropAllTables drops all the tables. Use with caution!
func DropAllTables() error {
	db := gdb.GetDB()

	if err := db.Migrator().DropTable(
		&userHobby{},
		&hobby{},
		&post{},
		&user{},
		&tempEmail{},
		&twoFABackup{},
		&twoFA{},
		&auth{},
	); err != nil {
		return err
	}

	fmt.Println("old tables are deleted!")
	return nil
}

// StartMigration automatically migrates all the tables.
//
//   - Only create tables with missing columns and missing indexes.
//   - Will not change/delete any existing columns and their types.
func StartMigration(configure gconfig.Configuration) error {
	db := gdb.GetDB()
	configureDB := configure.Database.RDBMS
	driver := configureDB.Env.Driver

	if driver == "mysql" {
		// db.Set() --> add table suffix during auto migration
		if err := db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
			&auth{},
			&twoFA{},
			&twoFABackup{},
			&tempEmail{},
			&user{},
			&post{},
			&hobby{},
		); err != nil {
			return err
		}

		fmt.Println("new tables are  migrated successfully!")
		return nil
	}

	if err := db.AutoMigrate(
		&auth{},
		&twoFA{},
		&twoFABackup{},
		&tempEmail{},
		&user{},
		&post{},
		&hobby{},
	); err != nil {
		return err
	}

	fmt.Println("new tables are  migrated successfully!")
	return nil
}
