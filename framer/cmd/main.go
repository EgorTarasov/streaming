package main

import (
	"context"

	"github.com/EgorTarasov/streaming/framer/internal"
	"github.com/rs/zerolog"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	defer cancel()
	if err := internal.Run(ctx); err != nil {
		panic(err)
	}
	// if err := internal.TestS3(ctx); err != nil {
	// 	panic(err)
	// }
}
