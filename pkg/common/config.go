package common

// Config defines all application-wide configuration options.
type Config struct {
	// Alias is the display name that everyone can see.
	Alias string

	// Hostname is the base endpoint for peers to connect to.
	Hostname string
}

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() Config {
	return Config{
		Alias: "Unnamed",
	}
}
