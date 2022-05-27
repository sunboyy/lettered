package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/sunboyy/lettered/pkg/management"
)

func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load configuration")
	}

	start(config)
}

func start(config Config) {
	managementAuth := management.NewAuth(config.Management)

	r := gin.Default()

	// Management APIs
	managementRouter := r.Group("/management")
	{
		managementHandler := &ManagementHandler{auth: managementAuth}

		managementRouter.POST("/login", managementHandler.Login)
	}

	r.Run()
}
