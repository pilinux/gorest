package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// Configuration - server and db configuration variables
type Configuration struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logger   LoggerConfig
}

// Config - load the configurations from .env
func Config() Configuration {
	var configuration Configuration

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.WithError(err).Panic("panic code: 101")
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
	serverTrustedIP := os.Getenv("TRUSTED_IP")

	loggerSentryDsn := os.Getenv("SentryDSN")

	accessKey := os.Getenv("ACCESS_KEY")
	accessKeyTTL, err := strconv.Atoi(os.Getenv("ACCESS_KEY_TTL"))
	if err != nil {
		log.WithError(err).Panic("panic code: 111")
	}
	refreshKey := os.Getenv("REFRESH_KEY")
	refreshKeyTTL, err := strconv.Atoi(os.Getenv("REFRESH_KEY_TTL"))
	if err != nil {
		log.WithError(err).Panic("panic code: 112")
	}

	hashPassMemory64, err := strconv.ParseUint((os.Getenv("HASHPASSMEMORY")), 10, 64)
	if err != nil {
		log.WithError(err).Panic("panic code: 121")
	}
	hashPassIterations64, err := strconv.ParseUint((os.Getenv("HASHPASSITERATIONS")), 10, 64)
	if err != nil {
		log.WithError(err).Panic("panic code: 122")
	}
	hashPassParallelism64, err := strconv.ParseUint((os.Getenv("HASHPASSPARALLELISM")), 10, 64)
	if err != nil {
		log.WithError(err).Panic("panic code: 123")
	}
	hashPassSaltLength64, err := strconv.ParseUint((os.Getenv("HASHPASSSALTLENGTH")), 10, 64)
	if err != nil {
		log.WithError(err).Panic("panic code: 124")
	}
	hashPassKeyLength64, err := strconv.ParseUint((os.Getenv("HASHPASSKEYLENGTH")), 10, 64)
	if err != nil {
		log.WithError(err).Panic("panic code: 125")
	}
	hashPassMemory := uint32(hashPassMemory64)
	hashPassIterations := uint32(hashPassIterations64)
	hashPassParallelism := uint8(hashPassParallelism64)
	hashPassSaltLength := uint32(hashPassSaltLength64)
	hashPassKeyLength := uint32(hashPassKeyLength64)

	configuration.Server.ServerPort = serverport
	configuration.Server.ServerEnv = serverEnv
	configuration.Server.ServerTrustedIP = serverTrustedIP

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
		log.WithError(err).Panic("panic code: 131")
	}
	configuration.Database.DbMaxOpenConns, err = strconv.Atoi(dbMaxOpenConns)
	if err != nil {
		log.WithError(err).Panic("panic code: 132")
	}
	configuration.Database.DbConnMaxLifetime, err = time.ParseDuration(dbConnMaxLifetime)
	if err != nil {
		log.WithError(err).Panic("panic code: 133")
	}
	configuration.Database.DbLogLevel, err = strconv.Atoi(dbLogLevel)
	if err != nil {
		log.WithError(err).Panic("panic code: 134")
	}

	configuration.Logger.SentryDsn = loggerSentryDsn

	configuration.Server.ServerJWT.AccessKey = accessKey
	configuration.Server.ServerJWT.AccessKeyTTL = accessKeyTTL
	configuration.Server.ServerJWT.RefreshKey = refreshKey
	configuration.Server.ServerJWT.RefreshKeyTTL = refreshKeyTTL

	configuration.Server.ServerHashPass.Memory = hashPassMemory
	configuration.Server.ServerHashPass.Iterations = hashPassIterations
	configuration.Server.ServerHashPass.Parallelism = hashPassParallelism
	configuration.Server.ServerHashPass.SaltLength = hashPassSaltLength
	configuration.Server.ServerHashPass.KeyLength = hashPassKeyLength

	return configuration
}
