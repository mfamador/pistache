// main package
package main

import (
	"github.com/mfamador/pistache/internal/config"
	_ "github.com/mfamador/pistache/internal/logger"
	"github.com/mfamador/pistache/internal/server"

	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Starting Pistache")

	// Start handling requests
	err := server.Start(config.Config.Server, &config.Config.Services)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start the HTTP server")
	}
}
