package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Configuration - server and db configuration variables
type Configuration struct {
	Server   ServerConfig
	Database DatabaseConfig
}

// Config - load the configurations from .env
func Config() Configuration {
	var configuration Configuration

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	dbDriver := os.Getenv("DBDRIVER")
	dbUser := os.Getenv("DBUSER")
	dbPass := os.Getenv("DBPASS")
	dbName := os.Getenv("DBNAME")
	dbHost := os.Getenv("DBHOST")
	dbport := os.Getenv("DBPORT")
	dbSslmode := os.Getenv("DBSSLMODE")
	dbTimeZone := os.Getenv("DBTIMEZONE")
	dbMaxIdleConns := os.Getenv("DBMAXIDLECONNS")
	dbMaxOpenConns := os.Getenv("DBMAXOPENCONNS")
	dbConnMaxLifetime := os.Getenv("DBCONNMAXLIFETIME")
	dbLogLevel := os.Getenv("DBLOGLEVEL")

	serverport := os.Getenv("APP_PORT")
	serverEnv := os.Getenv("APP_ENV")

	mySigningKey := os.Getenv("MySigningKey")
	JWTExpireTime, err := strconv.Atoi(os.Getenv("JWTExpireTime"))
	if err != nil {
		panic(err)
	}

	hashPassMemory64, err := strconv.ParseUint((os.Getenv("HASHPASSMEMORY")), 10, 64)
	if err != nil {
		panic(err)
	}
	hashPassIterations64, err := strconv.ParseUint((os.Getenv("HASHPASSITERATIONS")), 10, 64)
	if err != nil {
		panic(err)
	}
	hashPassParallelism64, err := strconv.ParseUint((os.Getenv("HASHPASSPARALLELISM")), 10, 64)
	if err != nil {
		panic(err)
	}
	hashPassSaltLength64, err := strconv.ParseUint((os.Getenv("HASHPASSSALTLENGTH")), 10, 64)
	if err != nil {
		panic(err)
	}
	hashPassKeyLength64, err := strconv.ParseUint((os.Getenv("HASHPASSKEYLENGTH")), 10, 64)
	if err != nil {
		panic(err)
	}
	hashPassMemory := uint32(hashPassMemory64)
	hashPassIterations := uint32(hashPassIterations64)
	hashPassParallelism := uint8(hashPassParallelism64)
	hashPassSaltLength := uint32(hashPassSaltLength64)
	hashPassKeyLength := uint32(hashPassKeyLength64)

	configuration.Server.ServerPort = serverport
	configuration.Server.ServerEnv = serverEnv

	configuration.Database.DbDriver = dbDriver
	configuration.Database.DbUser = dbUser
	configuration.Database.DbPass = dbPass
	configuration.Database.DbName = dbName
	configuration.Database.DbHost = dbHost
	configuration.Database.DbPort = dbport
	configuration.Database.DbSslmode = dbSslmode
	configuration.Database.DbTimeZone = dbTimeZone

	configuration.Database.DbMaxIdleConns, err = strconv.Atoi(dbMaxIdleConns)
	if err != nil {
		panic(err)
	}
	configuration.Database.DbMaxOpenConns, err = strconv.Atoi(dbMaxOpenConns)
	if err != nil {
		panic(err)
	}
	configuration.Database.DbConnMaxLifetime, err = time.ParseDuration(dbConnMaxLifetime)
	if err != nil {
		panic(err)
	}
	configuration.Database.DbLogLevel, err = strconv.Atoi(dbLogLevel)
	if err != nil {
		panic(err)
	}

	configuration.Server.ServerJWT.Key = mySigningKey
	configuration.Server.ServerJWT.Expire = JWTExpireTime

	configuration.Server.ServerHashPass.Memory = hashPassMemory
	configuration.Server.ServerHashPass.Iterations = hashPassIterations
	configuration.Server.ServerHashPass.Parallelism = hashPassParallelism
	configuration.Server.ServerHashPass.SaltLength = hashPassSaltLength
	configuration.Server.ServerHashPass.KeyLength = hashPassKeyLength

	return configuration
}
