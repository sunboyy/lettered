package p2p

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// NodeIDFromPubKey derives node ID from TLS certificate by extracting an ECDSA
// public key from the certificate and pass through NodeIDFromPubKey
func NodeIDFromCert(cert tls.Certificate) (string, error) {
	priv, ok := cert.PrivateKey.(*ecdsa.PrivateKey)
	if !ok {
		return "", errors.New("not an ecdsa private key")
	}
	return NodeIDFromPubKey(priv.Public())
}

// NodeIDFromPubKey derives node ID from ECDSA public key by applying SHA256
// to the PKIX-encoded format of the public key.
func NodeIDFromPubKey(pubKey any) (string, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", err
	}

	pubHash := sha256.Sum256(pubBytes)
	return hex.EncodeToString(pubHash[:]), nil
}

// CreateIdentifier creates a string representing the user's identifier by
// combining base64-encoded public key with hostname delimited by '@' sign.
func CreateIdentifier(nodeID string, hostname string) string {
	return fmt.Sprintf("%s@%s", nodeID, hostname)
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
