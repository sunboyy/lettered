package p2p

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/sunboyy/lettered/pkg/security"
)

// CreateIdentifier creates a string representing the user's identifier by
// combining base64-encoded public key with hostname delimited by '@' sign.
func CreateIdentifier(publicKey *btcec.PublicKey, hostname string) string {
	return fmt.Sprintf(
		"%s@%s",
		security.MarshalPublicKey(publicKey),
		hostname,
	)
}

// ExtractIndentifier extracts peer's identifier into public key and hostname by
// splitting with '@' sign. Zero or multiple '@' signs in the identifier is
// invalid.
func ExtractIdentifier(identifier string) (string, string, bool) {
	tokens := strings.Split(identifier, "@")
	if len(tokens) != 2 {
		return "", "", false
	}
	return tokens[0], tokens[1], true
}
