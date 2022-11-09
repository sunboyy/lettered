package p2p

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

const (
	EventPing         = "PING"
	EventFriendInvite = "FRIEND_INVITE"
)

// Peer is a wrapper of P2P client struct containing useful functionality for
// calling peer services. It contains hostname which can be used to construct
// an endpoint to call a peer.
type Peer struct {
	client *Client

	// identifier of the peer.
	identifier string
}

// NewPeer is a constructor of Peer.
func NewPeer(client *Client, identifier string) *Peer {
	return &Peer{
		client:     client,
		identifier: identifier,
	}
}

// Ping invokes PING event request.
func (p *Peer) Ping(req *PingRequest) (*PingResponse, error) {
	resBytes, err := p.client.Request(p.identifier, EventPing, req)
	if err != nil {
		return nil, err
	}

	var res PingResponse
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &res, nil
}

// FriendInvite invokes FRIEND_INVITE event request.
func (p *Peer) FriendInvite(req *FriendInviteRequest) (*FriendInviteResponse,
	error) {

	resBytes, err := p.client.Request(p.identifier, EventFriendInvite, req)
	if err != nil {
		return nil, err
	}

	var res FriendInviteResponse
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &res, nil
}
