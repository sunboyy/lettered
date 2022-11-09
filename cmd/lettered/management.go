package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/sunboyy/lettered/pkg/common"
	"github.com/sunboyy/lettered/pkg/friend"
	"github.com/sunboyy/lettered/pkg/management"
	"github.com/sunboyy/lettered/pkg/p2p"
)

var (
	// ErrUnauthorized is returned when the user does not request to the
	// management APIs with valid access token.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInvalidRequest is returned when the system cannot bind request
	// body or request query with the response struct.
	ErrInvalidRequest = errors.New("invalid request")

	// ErrInternalServerError is returned when the unexpected error occurs
	// while processing the request.
	ErrInternalServerError = errors.New("internal server error")
)

// ManagementHandler is a set of gin handlers functions that handles management
// functionality of the system.
type ManagementHandler struct {
	commonConfig  common.Config
	auth          *management.Auth
	friendManager *friend.Manager
	nodeID        string
}

// Middleware is an authentication middleware for the management APIs. It
// rejects all requests with no valid access token.
func (h *ManagementHandler) Middleware(ctx *gin.Context) {
	authHeader := ctx.Request.Header.Get("Authorization")
	authHeader = strings.TrimPrefix(authHeader, "Bearer ")
	if !h.auth.AccessTokenValid(authHeader) {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{"error": ErrUnauthorized.Error()},
		)
		ctx.Abort()
		return
	}

	ctx.Next()
}

// Login is a gin handler for logging in to the management console. It receives
// a password from the request body, generates a new access token if the
// given password is correct.
func (h *ManagementHandler) Login(ctx *gin.Context) {
	var req ManagementLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": ErrInvalidRequest.Error()},
		)
		return
	}

	accessToken, err := h.auth.Login(req.Password)
	if err != nil {
		if errors.Is(err, management.ErrIncorrectPassword) {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"error": err.Error()},
			)
		} else {
			log.Warn().Err(err).Msg("error processing login")
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{"error": ErrInternalServerError.Error()},
			)
		}
		return
	}

	ctx.JSON(http.StatusOK, ManagementLoginResponse{
		AccessToken: accessToken,
	})
}

// ManagementLoginRequest defines a request body of the management login API.
type ManagementLoginRequest struct {
	Password string `json:"password"`
}

// ManagementLoginResponse defines a response body of the management login API.
type ManagementLoginResponse struct {
	AccessToken string `json:"accessToken"`
}

// Identity is a gin handler returning the user's identifier for other people
// to connect.
func (h *ManagementHandler) Identity(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, IdentityResponse{
		Identifier: p2p.CreateIdentifier(
			h.nodeID,
			h.commonConfig.Hostname,
		),
	})
}

// IdentityResponse defines response body of the identity management API.
type IdentityResponse struct {
	Identifier string `json:"identifier"`
}

func (h *ManagementHandler) SendInvite(ctx *gin.Context) {
	var req SendInviteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": ErrInvalidRequest.Error()},
		)
		return
	}

	if err := h.friendManager.SendInvite(req.Identifier); err != nil {
		if errors.Is(err, friend.ErrInvalidIdentifier) ||
			errors.Is(err, friend.ErrInviteSelf) ||
			errors.Is(err, friend.ErrAlreadyFriend) {

			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"error": err.Error()},
			)
		} else {
			log.Warn().Err(err).Msg("error sending invite")
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{"error": ErrInternalServerError.Error()},
			)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

type SendInviteRequest struct {
	Identifier string `json:"identifier"`
}
