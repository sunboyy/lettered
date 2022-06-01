package main

import (
	"github.com/sunboyy/lettered/pkg/friend"
)

// PeerHandler contains a set of P2P handler functions receiving data from
// the wrapping function p2p.GinHandler.
type PeerHandler struct {
	friendManager *friend.Manager
}

// MyInfo is a P2P handler returning basic information that everyone can see.
func (h *PeerHandler) MyInfo(publicKey string, body []byte) (interface{}, error) {
	return h.friendManager.MyInfo(), nil
}
