package main

import (
	"context"
	"log"

	"github.com/EgorTarasov/streaming/framer/internal"
	"github.com/rs/zerolog"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	defer cancel()
	if err := internal.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
