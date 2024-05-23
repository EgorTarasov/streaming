package main

import (
	"context"

	"github.com/EgorTarasov/streaming/orchestrator/internal"
)

func main() {
	ctx := context.Background()
	if err := internal.Run(ctx); err != nil {
		panic(err)
	}
}
