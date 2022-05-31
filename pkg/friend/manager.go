package friend

import (
	"errors"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/sunboyy/lettered/pkg/common"
	"github.com/sunboyy/lettered/pkg/db"
	"github.com/sunboyy/lettered/pkg/p2p"
)

var (
	// ErrInvalidIdentifier is returned when identifier format is invalid.
	ErrInvalidIdentifier = errors.New("invalid identifier")
	// ErrAlreadyFriend is returned when sending friend request to peers
	// that are already the friend of the user.
	ErrAlreadyFriend = errors.New("already a friend")
)

// Manager contains a set of functionalities managing user's friends.
type Manager struct {
	commonConfig common.Config
	db           *db.DB
	p2pClient    *p2p.Client
}

// NewManager is a constructor of Manager.
func NewManager(commonConfig common.Config, db *db.DB,
	p2pClient *p2p.Client) *Manager {

	return &Manager{
		commonConfig: commonConfig,
		db:           db,
		p2pClient:    p2pClient,
	}
}

// MyInfo returns basic information about the user that anyone can see.
func (m *Manager) MyInfo() p2p.MyInfoResponse {
	return p2p.MyInfoResponse{
		Alias: m.commonConfig.Alias,
	}
}

// SendInvite sends friend request to the provided peer identifier. The
// identifier is a concatenation of public key and hostname delimited with an
// '@' sign.
func (m *Manager) SendInvite(identifier string) error {
	publicKey, hostname, err := extractIdentifier(identifier)
	if err != nil {
		return err
	}

	// Discard sending friend request if the peer is already a friend.
	alreadyFriend, err := m.db.FriendExists(publicKey)
	if err != nil {
		log.Warn().Str("source", "friend.Manager.SendInvite").
			Err(err).Msg("cannot request db.FriendExists")
		return err
	}
	if alreadyFriend {
		return ErrAlreadyFriend
	}

	// Send friend request to peer.
	res, err := p2p.NewPeer(m.p2pClient, hostname).
		Invite(p2p.InviteRequest{Hostname: m.commonConfig.Hostname})
	if err != nil {
		log.Warn().Str("source", "friend.Manager.SendInvite").
			Err(err).Msg("error sending invitation to peer")
		return err
	}

	// If peer accepts friend request, insert into friend database.
	if res.Accepted {
		return m.requestToFriend(&db.FriendRequest{
			PublicKey: publicKey,
			Hostname:  hostname,
		})
	}

	// Find previously created friend request in the database.
	friendReq, err := m.db.FindFriendRequest(publicKey)
	if err != nil {
		log.Warn().Str("source", "friend.Manager.SendInvite").
			Err(err).Msg("cannot request db.FindFriendRequest")
		return err
	}

	// If there is no previous friend request between the user and this
	// peer, create one.
	if friendReq == nil {
		_, err = m.db.CreateFriendRequest(
			publicKey,
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
func (m *Manager) ReceiveInvite(publicKey string, req p2p.InviteRequest) (
	p2p.InviteResponse, error) {

	// Immediately return if the requester is already a friend.
	alreadyFriend, err := m.db.FriendExists(publicKey)
	if err != nil {
		log.Warn().Str("source", "friend.Manager.ReceiveInvite").
			Err(err).Msg("cannot request db.FriendExists")
		return p2p.InviteResponse{}, err
	}
	if alreadyFriend {
		return p2p.InviteResponse{Accepted: true}, nil
	}

	// Find previously created friend request.
	friendReq, err := m.db.FindFriendRequest(publicKey)
	if err != nil {
		log.Warn().Str("source", "friend.Manager.ReceiveInvite").
			Err(err).Msg("cannot request db.FindFriendRequest")
		return p2p.InviteResponse{}, err
	}

	// If there is no previously created friend request, create one.
	if friendReq == nil {
		if _, err := m.db.CreateFriendRequest(
			publicKey,
			req.Hostname,
			false,
		); err != nil {
			log.Warn().
				Str("source", "friend.Manager.ReceiveInvite").
				Err(err).Msg("cannot request db.FriendExists")
			return p2p.InviteResponse{}, err
		}

		return p2p.InviteResponse{Accepted: false}, nil
	}

	// If the user has previously created a friend request to this peer
	// before, accept the friend request.
	if friendReq.IsInitiator {
		if err := m.requestToFriend(friendReq); err != nil {
			return p2p.InviteResponse{}, err
		}

		return p2p.InviteResponse{Accepted: true}, nil
	}

	// Update friend request to the latest value.
	friendReq.Hostname = req.Hostname
	if err := m.db.UpdateFriendRequest(friendReq); err != nil {
		log.Warn().Str("source", "friend.Manager.ReceiveInvite").
			Err(err).Msg("cannot request db.UpdateFriendRequest")
		return p2p.InviteResponse{}, err
	}

	return p2p.InviteResponse{Accepted: false}, nil
}

// requestToFriend converts friend request into friend. It fetches peer info API
// for getting necessary information and stores to the friend database.
// If there is a friend request previously created, this will delete it.
func (m *Manager) requestToFriend(friendReq *db.FriendRequest) error {
	peerInfo, err := p2p.NewPeer(m.p2pClient, friendReq.Hostname).MyInfo()
	if err != nil {
		log.Warn().Str("source", "friend.Manager.requestToFriend").
			Err(err).Msg("error getting peer's info")
		return err
	}

	if _, err := m.db.CreateFriend(friendReq, peerInfo.Alias); err != nil {
		log.Warn().Str("source", "friend.Manager.requestToFriend").
			Err(err).Msg("cannot request db.CreateFriend")
		return err
	}

	if err := m.db.DeleteFriendRequest(friendReq.PublicKey); err != nil {
		log.Warn().Str("source", "friend.Manager.requestToFriend").
			Err(err).Msg("cannot request DeleteFriendRequest")
		return err
	}

	return nil
}

// extractIndentifier extracts peer identifier into public key and hostname by
// splitting with '@' sign. Zero or multiple '@' signs in the identifier is
// invalid.
func extractIdentifier(identifier string) (string, string, error) {
	tokens := strings.Split(identifier, "@")
	if len(tokens) != 2 {
		return "", "", ErrInvalidIdentifier
	}
	return tokens[0], tokens[1], nil
}
