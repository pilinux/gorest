// main function of the example application
package main

import (
	"fmt"
	"net/http"
	"time"

	gconfig "github.com/pilinux/gorest/config"
	gdatabase "github.com/pilinux/gorest/database"
	gserver "github.com/pilinux/gorest/lib/server"
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
		sentryHook, err := middleware.NewSentryHook(
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
			if sentryHook != nil {
				defer func() {
					sentryHook.Flush(5 * time.Second)
				}()

				GoroutineLogger = log.New()
				GoroutineLogger.AddHook(sentryHook)
				// this log level is independent of the global log level
				GoroutineLogger.SetLevel(log.DebugLevel)
				GoroutineLogger.SetFormatter(&log.JSONFormatter{})

				GoroutineLogger.
					WithFields(log.Fields{
						"time": time.Now().Format(time.RFC3339),
						"ref":  "main",
					}).
					Debug("testing sentry integration in the main function")

			}
		}
		if GoroutineLogger == nil {
			fmt.Println("failed to create a logger for separate goroutines")
		}
		if GoroutineLogger != nil {
			if configure.Logger.SentryDsn != "" {
				// Example goroutine to keep logging periodically
				go func() {
					i := 0
					for {
						i++
						ref := fmt.Sprintf("goroutine - %d", i)
						fmt.Println("ref:", ref)

						fmt.Println("testing sentry integration in a separate goroutine")
						GoroutineLogger.
							WithFields(log.Fields{
								"time": time.Now().Format(time.RFC3339),
								"ref":  ref,
							}).
							Info("testing sentry integration in a separate goroutine")

						time.Sleep(5 * time.Second)
					}
				}()
			}
		}
	*/

	r, err := router.SetupRouter(configure)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Attaches the router to a http.Server
	srv := &http.Server{
		Addr:    configure.Server.ServerHost + ":" + configure.Server.ServerPort,
		Handler: r,
		// Add timeout to prevent Slowloris attacks
		ReadTimeout:       30 * time.Second, // max time to read the entire request including body
		ReadHeaderTimeout: 5 * time.Second,  // max time to read the request header
		WriteTimeout:      5 * time.Second,  // max time to generate and send the response
		IdleTimeout:       60 * time.Second, // important for keep-alive connections
	}

	// Start shutdown watcher
	const shutdownTimeout = 30 * time.Second
	done := make(chan struct{})
	// Wait for interrupt signal to gracefully shutdown the server
	go func() {
		err := gserver.GracefulShutdown(
			srv,
			shutdownTimeout,
			done,
			gdatabase.CloseAllDB,
		)
		if err != nil {
			fmt.Println(err)
		}
	}()

	// Start the server
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("server error: %v\n", err)
	}
	// Wait for the graceful shutdown to complete
	<-done
	fmt.Println("server shutdown complete")
}
