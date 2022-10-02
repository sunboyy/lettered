package friend

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/sunboyy/lettered/pkg/common"
	"github.com/sunboyy/lettered/pkg/db"
	"github.com/sunboyy/lettered/pkg/p2p"
)

var (
	// ErrInvalidIdentifier is returned when identifier format is invalid.
	ErrInvalidIdentifier = errors.New("invalid identifier")
	// ErrInviteSelf is return when the user is trying to invite himself.
	ErrInviteSelf = errors.New("cannot invite to self")
	// ErrAlreadyFriend is returned when sending friend request to peers
	// that are already the friend of the user.
	ErrAlreadyFriend = errors.New("already a friend")
)

// Manager contains a set of functionalities managing user's friends.
type Manager struct {
	commonConfig common.Config
	db           *db.DB
	p2pClient    *p2p.Client
	nodeID       string
}

// NewManager is a constructor of Manager.
func NewManager(commonConfig common.Config, db *db.DB,
	p2pClient *p2p.Client, nodeID string) *Manager {

	return &Manager{
		commonConfig: commonConfig,
		db:           db,
		p2pClient:    p2pClient,
		nodeID:       nodeID,
	}
}

// SendInvite sends friend request to the provided peer identifier. The
// identifier is a concatenation of node ID and hostname delimited with an
// '@' sign.
func (m *Manager) SendInvite(identifier string) error {
	nodeID, hostname, ok := p2p.ExtractIdentifier(identifier)
	if !ok {
		return ErrInvalidIdentifier
	}

	if nodeID == m.nodeID {
		return ErrInviteSelf
	}

	// Discard sending friend request if the peer is already a friend.
	alreadyFriend, err := m.db.FriendExists(nodeID)
	if err != nil {
		log.Warn().Str("source", "friend.Manager.SendInvite").
			Err(err).Msg("cannot request db.FriendExists")
		return err
	}
	if alreadyFriend {
		return ErrAlreadyFriend
	}

	// Send friend request to peer.
	res, err := p2p.NewPeer(m.p2pClient, identifier).
		FriendInvite(&p2p.FriendInviteRequest{
			Hostname: m.commonConfig.Hostname,
			Alias:    m.commonConfig.Alias,
		})
	if err != nil {
		log.Warn().Str("source", "friend.Manager.SendInvite").
			Err(err).Msg("error sending invitation to peer")
		return err
	}

	// If peer accepts friend request, insert into friend database.
	if res.Accepted {
		return m.requestToFriend(&db.FriendRequest{
			NodeID:   nodeID,
			Hostname: hostname,
		}, res.Alias)
	}

	// Find previously created friend request in the database.
	friendReq, err := m.db.FindFriendRequest(nodeID)
	if err != nil {
		log.Warn().Str("source", "friend.Manager.SendInvite").
			Err(err).Msg("cannot request db.FindFriendRequest")
		return err
	}

	// If there is no previous friend request between the user and this
	// peer, create one.
	if friendReq == nil {
		_, err = m.db.CreateFriendRequest(
			nodeID,
			hostname,
			true,
		)
		return err
	}

	// Update friend request to the latest value.
	friendReq.Hostname = hostname
	if !friendReq.IsInitiator {
		// Peer silently deleted the friend request to the user.
		// Perform as if the peer hasn't invited the user before.
		friendReq.IsInitiator = true
	}
	if err := m.db.UpdateFriendRequest(friendReq); err != nil {
		log.Warn().Str("source", "friend.Manager.SendInvite").
			Err(err).Msg("cannot request db.UpdateFriendRequest")
		return err
	}

	return nil
}

// ReceiveInvite processes invitation when the user receives friend requests.
// It inserts friend or friend request database depending on whether the user
// has previously invited this peer or not. If friend data is created, it will
// delete related friend request if any. If the user invite the same peer
// multiple times, it will update the previous friend request instead of
// creating a new one.
func (m *Manager) ReceiveInvite(nodeID string,
	req *p2p.FriendInviteRequest) (*p2p.FriendInviteResponse, error) {

	// Immediately return if the requester is already a friend.
	alreadyFriend, err := m.db.FriendExists(nodeID)
	if err != nil {
		log.Warn().Str("source", "friend.Manager.ReceiveInvite").
			Err(err).Msg("cannot request db.FriendExists")
		return nil, err
	}
	if alreadyFriend {
		return &p2p.FriendInviteResponse{
			Accepted: true,
			Alias:    m.commonConfig.Alias,
		}, nil
	}

	// Find previously created friend request.
	friendReq, err := m.db.FindFriendRequest(nodeID)
	if err != nil {
		log.Warn().Str("source", "friend.Manager.ReceiveInvite").
			Err(err).Msg("cannot request db.FindFriendRequest")
		return nil, err
	}

	// If there is no previously created friend request, create one.
	if friendReq == nil {
		if _, err := m.db.CreateFriendRequest(
			nodeID,
			req.Hostname,
			false,
		); err != nil {
			log.Warn().
				Str("source", "friend.Manager.ReceiveInvite").
				Err(err).Msg("cannot request db.FriendExists")
			return nil, err
		}

		return &p2p.FriendInviteResponse{Accepted: false}, nil
	}

	// If the user has previously created a friend request to this peer
	// before, accept the friend request.
	if friendReq.IsInitiator {
		if err := m.requestToFriend(friendReq, req.Alias); err != nil {
			return nil, err
		}

		return &p2p.FriendInviteResponse{
			Accepted: true,
			Alias:    m.commonConfig.Alias,
		}, nil
	}

	// Update friend request to the latest value.
	friendReq.Hostname = req.Hostname
	if err := m.db.UpdateFriendRequest(friendReq); err != nil {
		log.Warn().Str("source", "friend.Manager.ReceiveInvite").
			Err(err).Msg("cannot request db.UpdateFriendRequest")
		return nil, err
	}

	return &p2p.FriendInviteResponse{Accepted: false}, nil
}

// requestToFriend converts friend request into friend. It fetches peer info API
// for getting necessary information and stores to the friend database.
// If there is a friend request previously created, this will delete it.
func (m *Manager) requestToFriend(friendReq *db.FriendRequest,
	alias string) error {

	if _, err := m.db.CreateFriend(friendReq, alias); err != nil {
		log.Warn().Str("source", "friend.Manager.requestToFriend").
			Err(err).Msg("cannot request db.CreateFriend")
		return err
	}

	if err := m.db.DeleteFriendRequest(friendReq.NodeID); err != nil {
		log.Warn().Str("source", "friend.Manager.requestToFriend").
			Err(err).Msg("cannot request DeleteFriendRequest")
		return err
	}

	return nil
}
