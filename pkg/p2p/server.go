package p2p

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// errHeaderTooShort is an error indicating that the received message does not
// have enough length to be able to process the header.
var errHeaderTooShort = errors.New("header too short")

// HandlerFunc defines the handler used by P2P service.
type HandlerFunc func(nodeID string, body []byte) (protoreflect.ProtoMessage,
	error)

// Server is a custom TLS over TCP server that handles P2P communication from
// peer nodes.
type Server struct {
	cert       tls.Certificate
	port       int
	handlerMap map[string]HandlerFunc
}

// NewServer is the constructor function for Server.
func NewServer(cert tls.Certificate, port int) *Server {
	return &Server{
		cert:       cert,
		port:       port,
		handlerMap: map[string]HandlerFunc{},
	}
}

func (s *Server) On(event string, handler HandlerFunc) {
	s.handlerMap[event] = handler
}

// Run starts the server.
func (s *Server) Run() error {
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		Certificates:       []tls.Certificate{s.cert},
		ClientAuth:         tls.RequestClientCert,
		InsecureSkipVerify: true,
	}

	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(s.port), tlsConfig)
	if err != nil {
		return fmt.Errorf("tls listen: %w", err)
	}
	defer listener.Close()

	log.Info().Msgf("listening to p2p connection on port %d", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).
				Msg("error accepting connection")
			continue
		}

		log.Info().
			Msgf("accepted connection from %s", conn.RemoteAddr())

		tlsConn, ok := conn.(*tls.Conn)
		if !ok {
			log.Error().Err(err).
				Msg("could not cast net.Conn to *tls.Conn")
			continue
		}

		go s.handleConnection(tlsConn)
	}
}

// handleConnection is a middleware, authenticating the connection and transform
// request body for easier use in the handler functions. It rejects the client
// without certificate and then extract the request body as in the designed
// protocol format.
func (s *Server) handleConnection(conn *tls.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Error().Err(err).Msgf(
				"error closing connection from %s",
				conn.RemoteAddr())
		}
	}()

	// Handshake so that client certificate can be read from the server.
	if !conn.ConnectionState().HandshakeComplete {
		if err := conn.Handshake(); err != nil {
			log.Error().Err(err).Msg("error handshaking")
			return
		}
	}

	clientCerts := conn.ConnectionState().PeerCertificates
	if len(clientCerts) == 0 {
		log.Info().Msgf(
			"connection from %s does not provide certificate",
			conn.RemoteAddr(),
		)
		return
	}

	nodeID, err := NodeIDFromPubKey(clientCerts[0].PublicKey)
	if err != nil {
		log.Error().Err(err).Msg("error deriving node id")
		return
	}

	data, err := io.ReadAll(conn)
	if err != nil {
		log.Error().Err(err).Msg("error reading connection")
		return
	}

	header, bodyBytes, err := extractMessage(data)
	if err != nil {
		log.Warn().Err(err).Msg("unable to extract message")
		return
	}

	handler, ok := s.handlerMap[header.GetEvent()]
	if !ok {
		log.Debug().Msgf("no such route %s", header.GetEvent())
		return
	}

	response, err := handler(nodeID, bodyBytes)
	if err != nil {
		log.Warn().Err(err).Msg("handler error")
		return
	}

	responseBytes, err := proto.Marshal(response)
	if err != nil {
		log.Error().Err(err).Msgf("cannot encode response %v", response)
		return
	}
	if _, err := conn.Write(responseBytes); err != nil {
		log.Error().Err(err).Msgf("cannot encode response %v", response)
		return
	}
}

// extractMessage extracts request body to the protocol format
// [headerLength(2)||header(headerLength)||body(*)] the header is extracted
// using protobuf to get the event, while the body is remain untouched.
func extractMessage(data []byte) (*Header, []byte, error) {
	if len(data) < 2 {
		return nil, nil, errHeaderTooShort
	}

	headerLength := binary.BigEndian.Uint16(data[:2])

	if len(data) < int(2+headerLength) {
		return nil, nil, errHeaderTooShort
	}

	var header Header
	if err := proto.Unmarshal(data[2:2+headerLength], &header); err != nil {
		return nil, nil, fmt.Errorf("unmarshal header: %w", err)
	}

	return &header, data[2+headerLength:], nil
}
