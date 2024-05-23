package main

import (
	"context"

	"github.com/EgorTarasov/streaming/api/internal"
)

func main() {
	ctx := context.Background()
	if err := internal.Run(ctx); err != nil {
		panic(err)
	}
}
