package p2p

// Config defines the configuration options for communicating with peers.
type Config struct {
	// PrivateKey is a base64-encoded ECDSA private key. Public key derived
	// from this private key is your identity.
	PrivateKey string

	// ProxyURL is a URL of HTTP proxy that will be used when making
	// requests to peers. Usually, this will be a URL of a local Tor proxy
	// so that onion addresses can be used and ensures end-to-end
	// encryption.
	ProxyURL string
}

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() Config {
	return Config{
		PrivateKey: "",
		ProxyURL:   "socks5://127.0.0.1:9050",
	}
}
