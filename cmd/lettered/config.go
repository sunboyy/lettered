package main

import (
	"github.com/sunboyy/lettered/pkg/common"
	"github.com/sunboyy/lettered/pkg/db"
	"github.com/sunboyy/lettered/pkg/management"
	"github.com/sunboyy/lettered/pkg/p2p"
	"gopkg.in/ini.v1"
)

// Config contains all of the configuration options of the application.
type Config struct {
	Port       int
	Common     common.RawConfig
	DB         db.Config
	Management management.Config
	P2P        p2p.Config
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
		Port:       8080,
		Common:     common.DefaultConfig(),
		DB:         db.DefaultConfig(),
		Management: management.DefaultConfig(),
		P2P:        p2p.DefaultConfig(),
	}
}
