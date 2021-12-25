package config

// ServerConfig ...
type ServerConfig struct {
	ServerPort string // public port of server
	ServerEnv  string
	ServerJWT  struct {
		AccessKey     string
		RefreshKey    string
		AccessKeyTTL  int
		RefreshKeyTTL int
	}
	ServerHashPass struct {
		Memory      uint32
		Iterations  uint32
		Parallelism uint8
		SaltLength  uint32
		KeyLength   uint32
	}
	ServerTrustedIP string
}
