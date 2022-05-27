package management

// Config defines the configuration options for the management console.
type Config struct {
	// Password is the secret key for logging in to the management console.
	Password string

	// SessionTimeout is the duration (in seconds) that the session ID for
	// authenticaing to the management console can be used.
	SessionTimeout int
}

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() Config {
	return Config{
		Password:       "letteradm",
		SessionTimeout: 3600,
	}
}
