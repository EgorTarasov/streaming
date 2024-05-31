package internal

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/EgorTarasov/streaming/orchestrator/internal/config"
	"github.com/EgorTarasov/streaming/orchestrator/internal/controller"
	"github.com/EgorTarasov/streaming/orchestrator/internal/grpc/pb"
	"github.com/EgorTarasov/streaming/orchestrator/internal/grpc/server"
	"github.com/EgorTarasov/streaming/orchestrator/internal/repository/postgres"
	"github.com/EgorTarasov/streaming/orchestrator/pkg/db"
	"github.com/EgorTarasov/streaming/orchestrator/pkg/kafka"
	"github.com/EgorTarasov/streaming/orchestrator/pkg/minios3"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func Run(ctx context.Context) error {
	cfg := config.MustNew()

	pg, err := db.NewDb(ctx, &cfg.Postgres)
	if err != nil {
		return err
	}
	s3 := minios3.MustNew(&cfg.S3)
	// TODO: create topics required for service

	consumer := kafka.MustNewConsumer(&cfg.Consumer, false, time.Second)
	producer := kafka.MustNewAsyncProducer(&cfg.Consumer, sarama.NewRandomPartitioner, 0)
	// FIXME: read Only New messages from responses topic
	repo := postgres.New(pg)

	ctrl := controller.New(repo, s3, consumer, producer)
	s := server.New(ctrl)
	gs := grpc.NewServer()
	pb.RegisterVideoProcessingServiceServer(gs, s)
	go func() {
		log.Info().Msg("starting responses topic consumer")
		if er := ctrl.ProcessResultTopic(ctx); err != nil {
			panic(er)
		}
	}()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9999))
	if err != nil {
		return fmt.Errorf("failed to listen %v", err)
	}

	if err = gs.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
