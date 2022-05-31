package common

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/sunboyy/lettered/pkg/security"
)

// RawConfig defines raw application-wide configuration options. This struct is
// instantiated when the configuration file is read.
type RawConfig struct {
	// Alias is the display name that everyone can see.
	Alias string

	// Hostname is the base endpoint for peers to connect to.
	Hostname string

	// PrivateKey is a base64-encoded ECDSA private key. Public key
	// derived from this private key is the user's identity.
	PrivateKey string
}

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() RawConfig {
	return RawConfig{
		Alias: "Unnamed",
	}
}

// Config converts some of the fields in RawConfig struct into usable variable
// types.
func (c RawConfig) Config() (Config, error) {
	privateKey, err := security.ParsePrivateKey(c.PrivateKey)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Alias:      c.Alias,
		Hostname:   c.Hostname,
		PrivateKey: privateKey,
	}, nil
}

// Config defines all application-wide configuration options.
type Config struct {
	// Alias is the display name that everyone can see.
	Alias string

	// Hostname is the base endpoint for peers to connect to.
	Hostname string

	// PrivateKey is an ECDSA private key from the btcec library. Public key
	// derived from this private key is the user's identity.
	PrivateKey *btcec.PrivateKey
}
