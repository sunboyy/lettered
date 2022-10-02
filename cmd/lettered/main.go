package main

import (
	"crypto/tls"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/sunboyy/lettered/pkg/config"
	"github.com/sunboyy/lettered/pkg/db"
	"github.com/sunboyy/lettered/pkg/friend"
	"github.com/sunboyy/lettered/pkg/management"
	"github.com/sunboyy/lettered/pkg/p2p"
	"github.com/sunboyy/lettered/pkg/tlsutil"
)

func main() {
	cfg := config.LoadConfig()

	start(cfg)
}

func start(cfg config.Config) {
	cert, err := tlsutil.LoadOrGenerateCertificate(
		filepath.Join(cfg.AppDataDir, "tls.cert"),
		filepath.Join(cfg.AppDataDir, "tls.key"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load tls certificate")
	}

	nodeID, err := p2p.NodeIDFromCert(cert)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to derive node id from certificate")
	}

	db, err := db.Open(filepath.Join(cfg.AppDataDir, "db.sqlite"))
	if err != nil {
		panic(err)
	}

	p2pClient := p2p.NewClient(cert)
	friendManager := friend.NewManager(cfg.Common, db, p2pClient, nodeID)

	go startP2PServer(cert, cfg.P2PPort, friendManager)
	go startManagementServer(cfg, friendManager, nodeID)

	forever := make(chan struct{})
	<-forever
}

func startP2PServer(cert tls.Certificate, port int, friendManager *friend.Manager) {
	p2pServer := p2p.NewServer(cert, port)

	peerHandler := &PeerHandler{friendManager: friendManager}

	p2pServer.On(p2p.EventPing, peerHandler.Ping)
	p2pServer.On(p2p.EventFriendInvite, peerHandler.ReceiveInvite)

	if err := p2pServer.Run(); err != nil {
		log.Fatal().Err(err).Msg("error running p2p server")
	}
}

func startManagementServer(cfg config.Config, friendManager *friend.Manager, nodeID string) {
	managementAuth := management.NewAuth(cfg.Management)

	r := gin.Default()

	// Management APIs
	managementRouter := r.Group("/management")
	{
		managementHandler := &ManagementHandler{
			commonConfig:  cfg.Common,
			auth:          managementAuth,
			friendManager: friendManager,
			nodeID:        nodeID,
		}
		managementRouter.POST("/login", managementHandler.Login)

		managementRouter.Use(managementHandler.Middleware)
		managementRouter.GET("/identity", managementHandler.Identity)
		managementRouter.POST("/people/invite/send", managementHandler.SendInvite)
	}

	r.Run(":" + strconv.Itoa(cfg.Management.Port))
}
