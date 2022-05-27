package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
)

var (
	newHash = sha256.New
	random  = rand.Reader
	label   []byte
)

// ParsePrivateKey converts base64-encoded RSA private key into type
// *rsa.PrivateKey. If decoding base64 or parsing private key fails, the error
// will be returned.
func ParsePrivateKey(base64PrivateKey string) (*rsa.PrivateKey, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(base64PrivateKey)
	if err != nil {
		return nil, err
	}

	return x509.ParsePKCS1PrivateKey(privateKeyBytes)
}

// ParsePublicKey converts base64-encoded RSA public key into type
// *rsa.PublicKey. If decoding base64 or parsing public key fails, the error
// will be returned.
func ParsePublicKey(base64PublicKey string) (*rsa.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(base64PublicKey)
	if err != nil {
		return nil, err
	}

	return x509.ParsePKCS1PublicKey(publicKeyBytes)
}

// MarshalPublicKey encodes public key of type *rsa.PublicKey in x509 PKCS1
// format and then returns as a base64-encoded string.
func MarshalPublicKey(publicKey *rsa.PublicKey) string {
	publicKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	return base64.StdEncoding.EncodeToString(publicKeyBytes)
}

// Encrypt encrypts the data with the given public key and returns as
// base64-encoded string of the encrypted data. The data is encoded in JSON
// format before encryption.
func Encrypt(data interface{}, publicKey *rsa.PublicKey) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	encryptedData, err := rsa.EncryptOAEP(
		newHash(),
		random,
		publicKey,
		jsonData,
		label,
	)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedData), nil
}

// Decrypt decrypts the base64-encoded encrypted data with the given private key
// and returns a byte array of decrypted data. The decrypted data can be a
// JSON-formatted string and can be used to unmarshal into structs later.
func Decrypt(base64EncryptedData string, privateKey *rsa.PrivateKey) ([]byte,
	error) {

	encryptedData, err := base64.StdEncoding.DecodeString(base64EncryptedData)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptOAEP(
		newHash(),
		random,
		privateKey,
		encryptedData,
		label,
	)
}
