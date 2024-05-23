package internal

import (
	"context"

	"github.com/EgorTarasov/streaming/responser/internal/config"
	"github.com/EgorTarasov/streaming/responser/internal/controller"
	"github.com/EgorTarasov/streaming/responser/internal/kafka/reader"
	repo "github.com/EgorTarasov/streaming/responser/internal/repository/postgres"
	"github.com/EgorTarasov/streaming/responser/pkg/infrastructure/kafka"
	"github.com/EgorTarasov/streaming/responser/pkg/infrastructure/minios3"
	"github.com/EgorTarasov/streaming/responser/pkg/infrastructure/postgres"
	"github.com/rs/zerolog/log"
)

type Handler interface {
	SaveResult(ctx context.Context, message reader.PredictionMessage) error
}

func Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg := config.MustNew("config.yaml")

	db, err := postgres.NewDb(ctx, &cfg.Postgres)
	if err != nil {
		return err
	}

	s3 := minios3.MustNew(&cfg.S3)

	c, err := kafka.NewConsumer(kafka.Config{
		Brokers: []string{cfg.Kafka.Broker},
		Topic:   cfg.Kafka.ResultTopic,
	})
	if err != nil {
		return err
	}

	consumer := reader.New(c.Consumer, []string{cfg.Kafka.Broker}, cfg.Kafka.PredictTopic)
	repository := repo.NewPredictionRepository(db)
	ctrl := controller.NewController(&s3, "frames", repository)

	predictionMessages := make(chan reader.PredictionMessage)

	err = consumer.Subscribe(ctx, predictionMessages)
	if err != nil {
		return err
	}

	// TODO: save images into s3 and results into postgresql
	// TODO: create api
	log.Info().Msg("start consuming messages")
	for {
		select {
		case msg := <-predictionMessages:
			log.Info().Msg("got message")
			if err := ctrl.SaveResult(ctx, msg); err != nil {
				log.Info().Err(err).Msg("failed to save result")
			}
		}
	}

}
