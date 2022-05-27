package main

import (
	"github.com/sunboyy/lettered/pkg/management"
	"gopkg.in/ini.v1"
)

// Config contains all of the configuration options of the application.
type Config struct {
	Management management.Config
}

// LoadConfig instantiates a Config struct and fill in the configuration options
// from the config files in the following order:
//  1) ./config.ini
//
// The configuration option that is not provided in any of the config files will
// be set to default as described in the DefaultConfig function.
func LoadConfig() (Config, error) {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		return Config{}, err
	}

	config := DefaultConfig()
	if err := cfg.MapTo(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() Config {
	return Config{
		Management: management.DefaultConfig(),
	}
}