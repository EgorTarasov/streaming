package main

import (
	"context"

	"github.com/EgorTarasov/streaming/api/internal"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()
	log.Info().Msg("starting application")
	if err := internal.Run(ctx); err != nil {
		log.Info().Err(err).Msg("Error starting api service")
		panic(err)
	}
}
