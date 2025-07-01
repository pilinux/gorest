// main function of the example2 application.
package main

import (
	"fmt"
	"net/http"
	"time"

	gconfig "github.com/pilinux/gorest/config"
	gdb "github.com/pilinux/gorest/database"
	gserver "github.com/pilinux/gorest/lib/server"
	"github.com/qiniu/qmgo/options"

	"github.com/pilinux/gorest/example2/internal/database/migrate"
	"github.com/pilinux/gorest/example2/internal/router"
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
		// initialize RDBMS client
		for {
			if err := gdb.InitDB().Error; err != nil {
				fmt.Println(err)
				time.Sleep(10 * time.Second) // wait before retrying
				continue                     // retry initialization
			}
			break // exit loop if initialization is successful
		}

		// drop all tables from DB
		/*
			if err := migrate.DropAllTables(); err != nil {
				fmt.Println(err)
				return
			}
		*/

		// start DB migration
		if err := migrate.StartMigration(*configure); err != nil {
			fmt.Println(err)
			return
		}
	}

	if gconfig.IsRedis() {
		// initialize REDIS client
		for {
			if _, err := gdb.InitRedis(); err != nil {
				fmt.Println(err)
				time.Sleep(10 * time.Second) // wait before retrying
				continue                     // retry initialization
			}
			break // exit loop if initialization is successful
		}
	}

	if gconfig.IsMongo() {
		// initialize MONGO client
		for {
			if _, err := gdb.InitMongo(); err != nil {
				fmt.Println(err)
				time.Sleep(10 * time.Second) // wait before retrying
				continue                     // retry initialization
			}
			break // exit loop if initialization is successful
		}

		// example of dropping index "countryCode" from collection "geocodes" in database "map"
		/*
			indexes := []string{"countryCode"}
			if err := gdb.MongoDropIndex("map", "geocodes", indexes); err != nil {
				fmt.Println(err)
				return
			}
		*/

		// example of dropping all indexes from collection "geocodes" in database "map"
		/*
			if err := gdb.MongoDropAllIndexes("map", "geocodes"); err != nil {
				fmt.Println(err)
				return
			}
		*/

		// create new index for "countryCode" field
		index := options.IndexModel{
			Key: []string{"countryCode"},
		}
		// example of creating many indexes
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
		if err := gdb.MongoCreateIndex("map", "geocodes", index); err != nil {
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
				// example goroutine to keep logging periodically
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

	// attach the router to a http.Server
	srv := &http.Server{
		Addr:    configure.Server.ServerHost + ":" + configure.Server.ServerPort,
		Handler: r,
		// add timeout to prevent Slowloris attacks
		ReadTimeout:       30 * time.Second, // max time to read the entire request including body
		ReadHeaderTimeout: 5 * time.Second,  // max time to read the request header
		WriteTimeout:      5 * time.Second,  // max time to generate and send the response
		IdleTimeout:       60 * time.Second, // important for keep-alive connections
	}

	// start shutdown watcher
	const shutdownTimeout = 30 * time.Second
	done := make(chan struct{})
	// wait for interrupt signal to gracefully shutdown the server
	go func() {
		err := gserver.GracefulShutdown(
			srv,
			shutdownTimeout,
			done,
			gdb.CloseAllDB,
		)
		if err != nil {
			fmt.Println(err)
		}
	}()

	// start the server
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("server error: %v\n", err)
	}
	// wait for the graceful shutdown to complete
	<-done
	fmt.Println("server shutdown complete")
}
