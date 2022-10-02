package config

import (
	"os"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/rs/zerolog/log"
	"github.com/sunboyy/lettered/pkg/common"
	"github.com/sunboyy/lettered/pkg/management"
	"gopkg.in/ini.v1"
)

var defaultAppDataDir = btcutil.AppDataDir("lettered", false)

// Config contains all of the configuration options of the application.
type Config struct {
	AppDataDir string
	P2PPort    int
	Common     common.Config
	Management management.Config
}

// LoadConfig instantiates a Config struct and fill in the configuration options
// from the config files in the following order:
//  1. ./config.ini
//
// The configuration option that is not provided in any of the config files will
// be set to default as described in the DefaultConfig function.
func LoadConfig() Config {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatal().Err(err).Msg("config: cannot load config file")
	}

	config := DefaultConfig()
	if err := cfg.MapTo(&config); err != nil {
		log.Fatal().Err(err).Msg("config: cannot map config file to struct")
	}

	if err := os.MkdirAll(config.AppDataDir, 0700); err != nil {
		log.Fatal().Err(err).Msg("config: cannot ensure app data directory")
	}

	return config
}

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() Config {
	return Config{
		AppDataDir: defaultAppDataDir,
		P2PPort:    1926,
		Common:     common.DefaultConfig(),
		Management: management.DefaultConfig(),
	}
}
