package p2p

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"io"

	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
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
		return nil, errors.New("invalid identifier")
	}

	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		Certificates:       []tls.Certificate{c.cert},
		ClientAuth:         tls.RequestClientCert,
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", hostname, tlsConfig)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if len(conn.ConnectionState().PeerCertificates) == 0 {
		return nil, errors.New("the server does not provide a certificate")
	}

	actualNodeID, err := NodeIDFromPubKey(
		conn.ConnectionState().PeerCertificates[0].PublicKey,
	)
	if err != nil {
		return nil, err
	}

	if expectedNodeID != actualNodeID {
		return nil, errors.New("server sent invalid node id as expected")
	}

	header := Header{Event: event}
	headerBytes, err := proto.Marshal(&header)
	if err != nil {
		return nil, err
	}

	headerLength := make([]byte, 2)
	binary.BigEndian.PutUint16(headerLength, uint16(len(headerBytes)))

	bodyBytes, err := proto.Marshal(body)
	if err != nil {
		return nil, err
	}

	if _, err := conn.Write(headerLength); err != nil {
		return nil, err
	}
	if _, err := conn.Write(headerBytes); err != nil {
		return nil, err
	}
	if _, err := conn.Write(bodyBytes); err != nil {
		return nil, err
	}
	if err := conn.CloseWrite(); err != nil {
		return nil, err
	}

	return io.ReadAll(conn)
}
