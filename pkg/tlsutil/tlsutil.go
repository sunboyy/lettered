package tlsutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// LoadOrGenerateCertificate loads TLS key and certificate from the provided
// certFile and keyFile. If an error occurs while loading the certificate, the
// the new key and certificate will be issued.
func LoadOrGenerateCertificate(certFile, keyFile string) (tls.Certificate,
	error) {

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Info().Msg("tls: creating new certificate")
		return NewCertificate(certFile, keyFile)
	}
	return cert, nil
}

// NewCertificate creates a new TLS key and certificate and saves on the
// provided certFile and keyFile.
func NewCertificate(certFile, keyFile string) (tls.Certificate, error) {
	certBlock, keyBlock, err := generateCertificate()
	if err != nil {
		return tls.Certificate{}, err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("open cert file: %w", err)
	}
	if err := pem.Encode(certOut, certBlock); err != nil {
		return tls.Certificate{}, fmt.Errorf("encode cert pem: %w", err)
	}
	if err := certOut.Close(); err != nil {
		return tls.Certificate{}, fmt.Errorf("close cert file: %w", err)
	}

	keyOut, err := os.OpenFile(
		keyFile,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("open key file: %w", err)
	}
	if err := pem.Encode(keyOut, keyBlock); err != nil {
		return tls.Certificate{}, fmt.Errorf("encode key pem: %w", err)
	}
	if err := keyOut.Close(); err != nil {
		return tls.Certificate{}, fmt.Errorf("close key file: %w", err)
	}

	tlsCert, err := tls.X509KeyPair(
		pem.EncodeToMemory(certBlock),
		pem.EncodeToMemory(keyBlock),
	)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("x509 key pair: %w", err)
	}
	return tlsCert, nil
}

func generateCertificate() (*pem.Block, *pem.Block, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("random serial: %w", err)
	}

	notBefore := time.Now().Truncate(time.Hour * 24)
	notAfter := notBefore.Add(time.Hour * 24 * 365)
	keyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         "lettered",
			Organization:       []string{"Lettered"},
			OrganizationalUnit: []string{"Automatically Generated"},
		},
		DNSNames:           []string{"lettered"},
		NotBefore:          notBefore,
		NotAfter:           notAfter,
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		KeyUsage:           keyUsage,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
	}

	certDerBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		priv.Public(),
		priv,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create cert: %w", err)
	}

	privDerBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal privkey: %w", err)
	}

	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDerBytes,
	}
	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDerBytes,
	}

	return certBlock, keyBlock, nil
}
