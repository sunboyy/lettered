package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sunboyy/lettered/pkg/management"
)

type ManagementHandler struct {
	auth *management.Auth
}

func (h *ManagementHandler) Login(ctx *gin.Context) {
	var req ManagementLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unable to read request"})
		return
	}

	sessionID, ok := h.auth.Login(req.Password)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "incorrect password"})
		return
	}

	ctx.JSON(http.StatusOK, ManagementLoginResponse{
		SessionID: sessionID,
	})
}

type ManagementLoginRequest struct {
	Password string `json:"password"`
}

type ManagementLoginResponse struct {
	SessionID string `json:"sessionId"`
}