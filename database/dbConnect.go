// Package database handles connections to different
// types of databases.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pilinux/gorest/config"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// Import MySQL database driver
	// _ "github.com/jinzhu/gorm/dialects/mysql"
	"gorm.io/driver/mysql"

	// Import PostgreSQL database driver
	// _ "github.com/jinzhu/gorm/dialects/postgres"
	"gorm.io/driver/postgres"

	// Import SQLite3 database driver
	// _ "github.com/jinzhu/gorm/dialects/sqlite"
	"gorm.io/driver/sqlite"

	// Import Redis Driver
	"github.com/mediocregopher/radix/v4"

	// Import Mongo driver
	"github.com/qiniu/qmgo"
	"github.com/qiniu/qmgo/options"
	"go.mongodb.org/mongo-driver/event"
	opts "go.mongodb.org/mongo-driver/mongo/options"
)

// dbClient holds the gorm DB connection instance.
var dbClient *gorm.DB

var sqlDB *sql.DB
var err error

// redisClient holds the Redis client connection instance.
var redisClient radix.Client

// RedisConnTTL is the context deadline in seconds for Redis connections.
var RedisConnTTL int

// mongoClient holds the MongoDB client connection instance.
var mongoClient *qmgo.Client

// InitDB initializes the database connection.
func InitDB() *gorm.DB {
	var db = dbClient

	configureDB := config.GetConfig().Database.RDBMS

	driver := configureDB.Env.Driver
	username := configureDB.Access.User
	password := configureDB.Access.Pass
	database := configureDB.Access.DbName
	host := configureDB.Env.Host
	port := configureDB.Env.Port
	sslmode := configureDB.Ssl.Sslmode
	timeZone := configureDB.Env.TimeZone
	maxIdleConns := configureDB.Conn.MaxIdleConns
	maxOpenConns := configureDB.Conn.MaxOpenConns
	connMaxLifetime := configureDB.Conn.ConnMaxLifetime
	logLevel := configureDB.Log.LogLevel

	switch driver {
	case "mysql":
		address := host
		if port != "" {
			address += ":" + port
		}
		dsn := username + ":" + password + "@tcp(" + address + ")/" + database + "?charset=utf8mb4&parseTime=True&loc=Local"
		if sslmode == "" {
			sslmode = "disable"
		}
		if sslmode != "disable" {
			// use host machine's root CAs to verify
			if sslmode == "require" {
				dsn += "&tls=true"
			}

			// perform comprehensive SSL/TLS certificate validation using
			// certificate signed by a recognized CA or by a self-signed certificate
			if sslmode == "verify-ca" || sslmode == "verify-full" {
				dsn += "&tls=custom"
				err = InitTLSMySQL()
				if err != nil {
					db.Error = err
					return db
				}
			}
		}
		sqlDB, err = sql.Open(driver, dsn)
		if err != nil {
			err = fmt.Errorf("failed to open SQL connection: %w", err)
			db.Error = err
			return db
		}
		sqlDB.SetMaxIdleConns(maxIdleConns)       // max number of connections in the idle connection pool
		sqlDB.SetMaxOpenConns(maxOpenConns)       // max number of open connections in the database
		sqlDB.SetConnMaxLifetime(connMaxLifetime) // max amount of time a connection may be reused

		db, err = gorm.Open(mysql.New(mysql.Config{
			Conn: sqlDB,
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(logLevel)),
		})
		if err != nil {
			err = fmt.Errorf("failed to open GORM connection: %w", err)
			db.Error = err
			return db
		}
		// Only for debugging
		if err == nil {
			fmt.Println("DB connection successful!")
		}

	case "postgres":
		address := "host=" + host
		if port != "" {
			address += " port=" + port
		}
		dsn := address + " user=" + username + " dbname=" + database + " password=" + password + " TimeZone=" + timeZone
		if sslmode == "" {
			sslmode = "disable"
		}
		if sslmode != "disable" {
			if configureDB.Ssl.RootCA != "" {
				dsn += " sslrootcert=" + configureDB.Ssl.RootCA
			} else if configureDB.Ssl.ServerCert != "" {
				dsn += " sslrootcert=" + configureDB.Ssl.ServerCert
			}
			if configureDB.Ssl.ClientCert != "" {
				dsn += " sslcert=" + configureDB.Ssl.ClientCert
			}
			if configureDB.Ssl.ClientKey != "" {
				dsn += " sslkey=" + configureDB.Ssl.ClientKey
			}
		}
		dsn += " sslmode=" + sslmode

		sqlDB, err = sql.Open("pgx", dsn)
		if err != nil {
			err = fmt.Errorf("failed to open SQL connection: %w", err)
			db.Error = err
			return db
		}
		sqlDB.SetMaxIdleConns(maxIdleConns)       // max number of connections in the idle connection pool
		sqlDB.SetMaxOpenConns(maxOpenConns)       // max number of open connections in the database
		sqlDB.SetConnMaxLifetime(connMaxLifetime) // max amount of time a connection may be reused

		db, err = gorm.Open(postgres.New(postgres.Config{
			Conn: sqlDB,
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(logLevel)),
		})
		if err != nil {
			err = fmt.Errorf("failed to open GORM connection: %w", err)
			db.Error = err
			return db
		}
		// Only for debugging
		if err == nil {
			fmt.Println("DB connection successful!")
		}

	case "sqlite3":
		db, err = gorm.Open(sqlite.Open(database), &gorm.Config{
			Logger:                                   logger.Default.LogMode(logger.Silent),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			err = fmt.Errorf("failed to open GORM connection: %w", err)
			db.Error = err
			return db
		}
		// Only for debugging
		if err == nil {
			fmt.Println("DB connection successful!")
		}

	default:
		err = fmt.Errorf("the driver %s is not implemented yet", driver)
		db.Error = err
		return db
	}

	dbClient = db

	return dbClient
}

// GetDB returns the current database connection.
func GetDB() *gorm.DB {
	return dbClient
}

// InitRedis initializes the Redis client connection.
func InitRedis() (radix.Client, error) {
	configureRedis := config.GetConfig().Database.REDIS

	RedisConnTTL = configureRedis.Conn.ConnTTL
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(RedisConnTTL)*time.Second)
	defer cancel()

	rClient, err := (radix.PoolConfig{
		Size: configureRedis.Conn.PoolSize,
	}).New(ctx, "tcp", fmt.Sprintf("%v:%v",
		configureRedis.Env.Host,
		configureRedis.Env.Port))
	if err != nil {
		err = fmt.Errorf("failed to connect to Redis: %w", err)
		return rClient, err
	}
	// Only for debugging
	fmt.Println("REDIS pool connection successful!")

	redisClient = rClient

	return redisClient, nil
}

// GetRedis returns the current Redis client connection.
func GetRedis() radix.Client {
	return redisClient
}

// InitMongo initializes the MongoDB client connection.
func InitMongo() (*qmgo.Client, error) {
	configureMongo := config.GetConfig().Database.MongoDB

	// Connect to the database or cluster
	uri := configureMongo.Env.URI

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(configureMongo.Env.ConnTTL)*time.Second)
	defer cancel()

	clientConfig := &qmgo.Config{
		Uri:         uri,
		MaxPoolSize: &configureMongo.Env.PoolSize,
	}
	serverAPIOptions := opts.ServerAPI(opts.ServerAPIVersion1)

	opt := opts.Client().SetAppName(configureMongo.Env.AppName)
	opt.SetServerAPIOptions(serverAPIOptions)

	// for monitoring pool
	if configureMongo.Env.PoolMon == "yes" {
		poolMonitor := &event.PoolMonitor{
			Event: func(evt *event.PoolEvent) {
				switch evt.Type {
				case event.GetSucceeded:
					fmt.Println("GetSucceeded")
				case event.ConnectionReturned:
					fmt.Println("ConnectionReturned")
				}
			},
		}
		opt.SetPoolMonitor(poolMonitor)
	}

	client, err := qmgo.NewClient(ctx, clientConfig, options.ClientOptions{ClientOptions: opt})
	if err != nil {
		return client, err
	}

	// Only for debugging
	fmt.Println("MongoDB pool connection successful!")

	mongoClient = client

	return mongoClient, nil
}

// GetMongo returns the current MongoDB client connection.
func GetMongo() *qmgo.Client {
	return mongoClient
}

// CloseAllDB closes all database connections.
func CloseAllDB() error {
	err := CloseSQL()
	if err != nil {
		return err
	}

	err = CloseRedis()
	if err != nil {
		return err
	}

	err = CloseMongo()
	if err != nil {
		return err
	}

	return nil
}

// CloseSQL closes the SQL database connection.
func CloseSQL() error {
	if dbClient != nil {
		fmt.Println("Closing SQL DB connection...")
		db, err := dbClient.DB()
		if err != nil {
			return err
		}
		if err := db.Close(); err != nil {
			return err
		}
		dbClient = nil
		sqlDB = nil
		fmt.Println("SQL DB connection closed!")
	}

	return nil
}

// CloseRedis closes the Redis database connection.
func CloseRedis() error {
	if redisClient != nil {
		fmt.Println("Closing Redis connection...")
		if err := radix.Client.Close(redisClient); err != nil {
			return err
		}
		redisClient = nil
		fmt.Println("Redis connection closed!")
	}

	return nil
}

// CloseMongo closes the MongoDB database connection.
func CloseMongo() error {
	if mongoClient != nil {
		fmt.Println("Closing MongoDB connection...")
		if err := mongoClient.Close(context.Background()); err != nil {
			return err
		}
		mongoClient = nil
		fmt.Println("MongoDB connection closed!")
	}

	return nil
}
