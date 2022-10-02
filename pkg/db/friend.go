package db

import (
	"errors"

	"gorm.io/gorm"
)

// FriendRequest is a temporary data structure for unaccepted friend requests
// (both incoming and outgoing requests).
type FriendRequest struct {
	gorm.Model

	// NodeID is an identity of the peer with which the request has been made.
	NodeID string

	// Hostname is a host endpoint of the peer that you can communicate to.
	Hostname string

	// IsInitiator is a boolean flag representing whether the friend request
	// is initiated by own or by peer.
	IsInitiator bool
}

// FindFriendRequest returns a friend request with the specified public key.
// An error will not be returned if there is no record found the first return
// value will be nil.
func (db *DB) FindFriendRequest(nodeID string) (*FriendRequest,
	error) {

	var friendRequest FriendRequest
	result := db.backend.Where("node_id = ?", nodeID).First(&friendRequest)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &friendRequest, nil
}

// CreateFriendRequest inserts a friend request to the database.
func (db *DB) CreateFriendRequest(nodeID string, hostname string,
	isInitiator bool) (*FriendRequest, error) {

	friendReq := FriendRequest{
		NodeID:      nodeID,
		Hostname:    hostname,
		IsInitiator: isInitiator,
	}
	result := db.backend.Create(&friendReq)
	if result.Error != nil {
		return nil, result.Error
	}

	return &friendReq, nil
}

// UpdateFriendRequest updates a friend request to the database by reading
// information in the friend request struct.
func (db *DB) UpdateFriendRequest(friendReq *FriendRequest) error {
	result := db.backend.Save(friendReq)
	return result.Error
}

// DeleteFriendRequest deletes friend request that matches the specified public
// key.
func (db *DB) DeleteFriendRequest(nodeID string) error {
	result := db.backend.Where("node_id = ?", nodeID).
		Delete(&FriendRequest{})
	return result.Error
}

// Friend is a data structure for friends that are already accepted both ways.
type Friend struct {
	gorm.Model
	NodeID   string
	Hostname string
	Alias    string
}

// CreateFriend inserts a new friend data into the friend database using the
// information from friend request struct.
func (db *DB) CreateFriend(friendReq *FriendRequest, alias string) (*Friend,
	error) {

	friend := Friend{
		NodeID:   friendReq.NodeID,
		Hostname: friendReq.Hostname,
		Alias:    alias,
	}
	result := db.backend.Create(&friend)
	if result.Error != nil {
		return nil, result.Error
	}

	return &friend, nil
}

// FriendExists checks whether there is a friend with the specified public key
// stored in the database.
func (db *DB) FriendExists(nodeID string) (bool, error) {
	var count int64
	result := db.backend.Model(&Friend{}).
		Where("node_id = ?", nodeID).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}
