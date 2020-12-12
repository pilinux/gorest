package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Configuration - server and db configuration variables
type Configuration struct {
	Server   ServerConfig
	Database DatabaseConfig
}

// ConfigMain - load the configurations from .env
func ConfigMain() Configuration {
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
	serverport := os.Getenv("APP_PORT")
	serverEnv := os.Getenv("APP_ENV")

	configuration.Server.ServerPort = serverport
	configuration.Server.ServerEnv = serverEnv
	configuration.Database.DbDriver = dbDriver
	configuration.Database.DbUser = dbUser
	configuration.Database.DbPass = dbPass
	configuration.Database.DbName = dbName
	configuration.Database.DbHost = dbHost
	configuration.Database.DbPort = dbport

	return configuration
}
