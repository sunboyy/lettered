package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/sunboyy/lettered/pkg/db"
	"github.com/sunboyy/lettered/pkg/friend"
	"github.com/sunboyy/lettered/pkg/management"
	"github.com/sunboyy/lettered/pkg/p2p"
)

func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load configuration")
	}

	start(config)
}

func start(config Config) {
	db, err := db.New(config.DB)
	if err != nil {
		panic(err)
	}

	commonConfig, err := config.Common.Config()
	if err != nil {
		panic(err)
	}

	p2pClient, err := p2p.NewClient(commonConfig, config.P2P)
	if err != nil {
		panic(err)
	}
	friendManager := friend.NewManager(commonConfig, db, p2pClient)

	managementAuth := management.NewAuth(config.Management)

	r := gin.Default()

	// Management APIs
	managementRouter := r.Group("/management")
	{
		managementHandler := &ManagementHandler{
			commonConfig:  commonConfig,
			auth:          managementAuth,
			friendManager: friendManager,
		}
		managementRouter.POST("/login", managementHandler.Login)

		managementRouter.Use(managementHandler.Middleware)
		managementRouter.GET("/identity", managementHandler.Identity)
		managementRouter.GET("/people/peer-info", managementHandler.PeerInfo)
	}

	// P2P communication APIs
	peerRouter := r.Group("/peer")
	{
		peerHandler := &PeerHandler{friendManager: friendManager}
		peerRouter.GET("/people/my-info", p2p.GinHandler(peerHandler.MyInfo))
	}

	r.Run(":" + strconv.Itoa(config.Port))
}
