package config

// ServerConfig - public port of server
type ServerConfig struct {
	ServerPort string
	ServerEnv  string
	ServerJWT  struct {
		Key    string
		Expire int
	}
}
