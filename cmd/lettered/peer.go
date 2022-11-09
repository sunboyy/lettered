package main

import (
	"fmt"

	"github.com/sunboyy/lettered/pkg/friend"
	"github.com/sunboyy/lettered/pkg/p2p"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// PeerHandler contains a set of P2P handler functions.
type PeerHandler struct {
	friendManager *friend.Manager
}

func (h *PeerHandler) Ping(nodeID string, body []byte) (
	protoreflect.ProtoMessage, error) {

	var req p2p.PingRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("unmarshal req body: %w", err)
	}

	return &p2p.PingResponse{
		Message: req.GetMessage(),
	}, nil
}

func (h *PeerHandler) ReceiveInvite(nodeID string, body []byte) (
	protoreflect.ProtoMessage, error) {

	var req p2p.FriendInviteRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("unmarshal req body: %w", err)
	}

	res, err := h.friendManager.ReceiveInvite(nodeID, &req)
	if err != nil {
		return nil, fmt.Errorf("fm receive invite: %w", err)
	}
	return res, nil
}
