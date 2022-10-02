package management

// Config defines the configuration options for the management console.
type Config struct {
	// Port specifies the HTTP port for the management server API.
	Port int

	// Password is the secret key for logging in to the management console.
	Password string

	// SessionTimeout is the duration (in seconds) that the access token for
	// authenticaing to the management console can be used.
	SessionTimeout int
}

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() Config {
	return Config{
		Port:           11926,
		Password:       "letteradm",
		SessionTimeout: 3600,
	}
}
