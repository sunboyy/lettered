package p2p

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	// ErrStatusNotOK is returned when calling peers and not returning the
	// response with status code 200 (OK).
	ErrStatusNotOK = errors.New("status code is not 200")
)

// Peer is a wrapper of P2P client struct containing useful functionality for
// calling peer services. It contains hostname which can be used to construct
// an endpoint to call a peer.
type Peer struct {
	client *Client

	// hostname of the peer.
	hostname string
}

// NewPeer is a constructor of Peer.
func NewPeer(client *Client, hostname string) *Peer {
	return &Peer{
		client:   client,
		hostname: hostname,
	}
}

// MyInfo requests to peer to get information about the peer user.
func (p *Peer) MyInfo() (MyInfoResponse, error) {
	res, err := p.client.GET(p.hostname, "/peer/people/my-info")
	if err != nil {
		return MyInfoResponse{}, err
	}

	if res.StatusCode != http.StatusOK {
		return MyInfoResponse{}, ErrStatusNotOK
	}

	var data MyInfoResponse
	err = json.NewDecoder(res.Body).Decode(&data)

	return data, err
}

// MyInfoResponse defines the response body when calling (*Peer).MyInfo.
type MyInfoResponse struct {
	// Alias is the peer's alias.
	Alias string `json:"alias"`
}

// Invite requests to peer to invite the peer to be a friend.
func (p *Peer) Invite(req InviteRequest) (InviteResponse, error) {
	res, err := p.client.POST(
		p.hostname,
		"/peer/people/invite/receive",
		req,
	)
	if err != nil {
		return InviteResponse{}, err
	}

	if res.StatusCode != http.StatusOK {
		return InviteResponse{}, ErrStatusNotOK
	}

	var data InviteResponse
	err = json.NewDecoder(res.Body).Decode(&data)

	return data, err
}

// InviteRequest defines the request body when calling (*Peer).Invite.
type InviteRequest struct {
	// Hostname of the user. When peers read this field, they will know how
	// to send invitation response back.
	Hostname string `json:"hostname"`
}

// InviteResponse defines the response body when calling (*Peer).Invite.
type InviteResponse struct {
	// Accepted indicates whether the peer accepts the invitation request.
	// It can be true if the user is already a friend of the peer of the
	// peer has previously invited the user to be a friend.
	Accepted bool `json:"accepted"`
}
