package config

// ServerConfig holds server configuration.
type ServerConfig struct {
	ServerHost string
	ServerPort string // public port of server
	ServerEnv  string
}
