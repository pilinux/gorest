package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Configuration struct {
	Server   ServerConfig
	Database DatabaseConfig
}

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
	serverport := os.Getenv("SERVERPORT")

	configuration.Server.ServerPort = serverport
	configuration.Database.DbDriver = dbDriver
	configuration.Database.DbUser = dbUser
	configuration.Database.DbPass = dbPass
	configuration.Database.DbName = dbName
	configuration.Database.DbHost = dbHost
	configuration.Database.DbPort = dbport

	return configuration
}
