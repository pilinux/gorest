// main function of the example application
package main

import (
	"fmt"

	gconfig "github.com/pilinux/gorest/config"
	gdatabase "github.com/pilinux/gorest/database"
	"github.com/qiniu/qmgo/options"

	"github.com/pilinux/gorest/example/database/migrate"
	"github.com/pilinux/gorest/example/router"
)

func main() {
	// set configs
	err := gconfig.Config()
	if err != nil {
		fmt.Println(err)
		return
	}

	// read configs
	configure := gconfig.GetConfig()

	if gconfig.IsRDBMS() {
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

		// Manually set foreign key for MySQL and PostgreSQL
		if err := migrate.SetPkFk(); err != nil {
			fmt.Println(err)
			return
		}
	}

	if gconfig.IsRedis() {
		// Initialize REDIS client
		if _, err := gdatabase.InitRedis(); err != nil {
			fmt.Println(err)
			return
		}
	}

	if gconfig.IsMongo() {
		// Initialize MONGO client
		if _, err := gdatabase.InitMongo(); err != nil {
			fmt.Println(err)
			return
		}

		// Example of dropping index "countryCode" from collection "geocodes" in database "map"
		/*
			indexes := []string{"countryCode"}
			if err := gdatabase.MongoDropIndex("map", "geocodes", indexes); err != nil {
				fmt.Println(err)
				return
			}
		*/

		// Example of dropping all indexes from collection "geocodes" in database "map"
		/*
			if err := gdatabase.MongoDropAllIndexes("map", "geocodes"); err != nil {
				fmt.Println(err)
				return
			}
		*/

		// Create new index for "countryCode" field
		index := options.IndexModel{
			Key: []string{"countryCode"},
		}
		// Example of creating many indexes
		/*
			indexes := []options.IndexModel{
				{
					Key: []string{"state"},
				},
				{
					Key: []string{"countryCode"},
				},
			}
		*/
		if err := gdatabase.MongoCreateIndex("map", "geocodes", index); err != nil {
			fmt.Println(err)
			return
		}
	}

	// example of using sentry in separate goroutines
	/*
		var GoroutineLogger *log.Logger
		sentryHook, err := middleware.InitSentry(
			configure.Logger.SentryDsn,
			configure.Server.ServerEnv,
			configure.Version,
			configure.Logger.PerformanceTracing,
			configure.Logger.TracesSampleRate,
		)
		if err != nil {
			fmt.Println(err)
		}
		if err == nil {
			sentryHook.SetFlushTimeout(5 * time.Second)
			defer sentryHook.Flush()
			GoroutineLogger = log.New()
			GoroutineLogger.AddHook(sentryHook)
		}
		if GoroutineLogger == nil {
			fmt.Println("failed to create a logger for separate goroutines")
		}
		if GoroutineLogger != nil {
			if configure.Logger.SentryDsn != "" {
				i := 0
				for {
					i++
					ref := fmt.Sprintf("goroutine - %d", i)
					fmt.Println("ref:", ref)
					go func() {
						fmt.Println("testing sentry integration in a separate goroutine")
						GoroutineLogger.
							WithFields(log.Fields{
								"time": time.Now().Format(time.RFC3339),
								"ref":  ref,
							}).
							Info("testing sentry integration in a separate goroutine")
					}()

					time.Sleep(5 * time.Second)
				}
			}
		}
	*/

	r, err := router.SetupRouter(configure)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Attaches the router to a http.Server and starts listening and serving HTTP requests
	err = r.Run(configure.Server.ServerHost + ":" + configure.Server.ServerPort)
	if err != nil {
		fmt.Println(err)
		return
	}
}
