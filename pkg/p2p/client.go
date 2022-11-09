package p2p

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

var (
	// errNoServerCert is an error indicating that the server does not
	// provide a certificate to verify themselves.
	errNoServerCert = errors.New("server does not provide a certificate")

	// errUnexpectedServerNodeID is an error indicating that the node ID
	// derived from the server public's key does not match with the
	// expected one in the identifier.
	errUnexpectedServerNodeID = errors.New("unexpected server node id")

	// errInvalidIdentifier is an error indicating that the identifier
	// cannot be extracted because of an unexpected pattern.
	errInvalidIdentifier = errors.New("invalid identifier")
)

// Client is an P2P client to communicate with peers. Communication is performed
// over TCP with TLS layer to ensure confidentiality and integrity. The
// application layer is customized to allow verification of peers without the
// need of certificate authorities.
type Client struct {
	// cert is a client TLS certificate used for authenticating server
	// certificate request.
	cert tls.Certificate
}

// NewClient is a constructor function for Client.
func NewClient(cert tls.Certificate) *Client {
	return &Client{
		cert: cert,
	}
}

// Request sends a P2P request to the specified identifier. After dialing to
// a peer, it checks if the server node ID is the same as stated in the
// identifier and rejects connection with invalid certificate. After
// authentication succeeds, it constructs and sends a message to the server
// in the protocol format [headerLength(2)||header(headerLength)||body(*)].
func (c *Client) Request(identifier string, event string,
	body protoreflect.ProtoMessage) ([]byte, error) {

	expectedNodeID, hostname, ok := ExtractIdentifier(identifier)
	if !ok {
		return nil, errInvalidIdentifier
	}

	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		Certificates:       []tls.Certificate{c.cert},
		ClientAuth:         tls.RequestClientCert,
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", hostname, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	if len(conn.ConnectionState().PeerCertificates) == 0 {
		return nil, errNoServerCert
	}

	actualNodeID, err := NodeIDFromPubKey(
		conn.ConnectionState().PeerCertificates[0].PublicKey,
	)
	if err != nil {
		return nil, err
	}

	if expectedNodeID != actualNodeID {
		return nil, errUnexpectedServerNodeID
	}

	header := Header{Event: event}
	headerBytes, err := proto.Marshal(&header)
	if err != nil {
		return nil, fmt.Errorf("marshal header proto: %w", err)
	}

	headerLength := make([]byte, 2)
	binary.BigEndian.PutUint16(headerLength, uint16(len(headerBytes)))

	bodyBytes, err := proto.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body proto: %w", err)
	}

	if _, err := conn.Write(headerLength); err != nil {
		return nil, fmt.Errorf("write header length: %w", err)
	}
	if _, err := conn.Write(headerBytes); err != nil {
		return nil, fmt.Errorf("write header: %w", err)
	}
	if _, err := conn.Write(bodyBytes); err != nil {
		return nil, fmt.Errorf("write body: %w", err)
	}
	if err := conn.CloseWrite(); err != nil {
		return nil, fmt.Errorf("close write: %w", err)
	}

	response, err := io.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return response, nil
}
