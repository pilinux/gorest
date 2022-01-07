package config

// SecurityConfig ...
type SecurityConfig struct {
	BasicAuth struct {
		Username string
		Password string
	}
	JWT struct {
		AccessKey     string
		RefreshKey    string
		AccessKeyTTL  int
		RefreshKeyTTL int
	}
	HashPass struct {
		Memory      uint32
		Iterations  uint32
		Parallelism uint8
		SaltLength  uint32
		KeyLength   uint32
	}
	Firewall struct {
		ListType string
		IP       string
	}
	TrustedIP string
}
