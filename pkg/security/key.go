package security

import (
	"encoding/base64"

	"github.com/btcsuite/btcd/btcec/v2"
)

// ParsePrivateKey constructs *btcec.PrivateKey from the base64-encoded string
// of ECDSA private key byte array data. The error will be returned if decoding
// base64 fails.
func ParsePrivateKey(privateKeyString string) (*btcec.PrivateKey, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(
		privateKeyString,
	)
	if err != nil {
		return nil, err
	}

	privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)
	return privateKey, nil
}

// ParsePublicKey constructs *btcec.PublicKey from the base64-encoded string of
// ECDSA public key byte array data. The error will be returned if decoding
// base64 fails.
func ParsePublicKey(publicKeyString string) (*btcec.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyString)
	if err != nil {
		return nil, err
	}

	return btcec.ParsePubKey(publicKeyBytes)
}

// MarshalPrivateKey serializes private key byte array data into base64-encoded
// string.
func MarshalPrivateKey(privateKey *btcec.PrivateKey) string {
	privateKeyBytes := privateKey.Serialize()
	return base64.StdEncoding.EncodeToString(privateKeyBytes)
}

// MarshalPublicKey serializes public key byte array data into base64-encoded
// string. Serialization uses compressed Bitcoin private key format so that
// it takes less network resource to transmit though the network.
func MarshalPublicKey(publicKey *btcec.PublicKey) string {
	publicKeyBytes := publicKey.SerializeCompressed()
	return base64.StdEncoding.EncodeToString(publicKeyBytes)
}
