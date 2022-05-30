package db

// Config defines all configuration options for the data storage.
type Config struct {
	DBPath string
}

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() Config {
	return Config{
		DBPath: "",
	}
}
