package config

// LoggerConfig holds logger configuration.
type LoggerConfig struct {
	Activate           string
	SentryDsn          string
	PerformanceTracing string
	TracesSampleRate   string
}
