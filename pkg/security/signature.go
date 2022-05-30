package security

import (
	"encoding/base64"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/rs/zerolog/log"
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

// VerifySignature verifies base64-encoded ECDSA signature with provided public
// key and data hash. It returns true only if there is no error decoding the
// signature string and the signature is valid with the provided public key and
// hash.
func VerifySignature(publicKey *btcec.PublicKey, dataHash []byte,
	signatureString string) bool {

	signatureBytes, err := base64.StdEncoding.DecodeString(
		signatureString,
	)
	if err != nil {
		log.Debug().Str("source", "security.VerifySignature").
			Err(err).Msg("cannot decode base64 signature")
		return false
	}

	signature, err := ecdsa.ParseSignature(signatureBytes)
	if err != nil {
		log.Debug().Str("source", "security.VerifySignature").
			Err(err).Msg("cannot parse ECDSA signature")
		return false
	}

	return signature.Verify(dataHash, publicKey)
}
