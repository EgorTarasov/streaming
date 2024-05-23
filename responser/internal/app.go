package internal

import (
	"context"

	"github.com/EgorTarasov/streaming/responser/internal/config"
	"github.com/EgorTarasov/streaming/responser/internal/kafka/reader"
	"github.com/EgorTarasov/streaming/responser/pkg/infrastructure/kafka"
	"github.com/EgorTarasov/streaming/responser/pkg/infrastructure/postgres"
)

type Handler interface {
	SaveResult(ctx context.Context, message reader.PredictionMessage) error
}

func Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg := config.MustNew()

	db, err := postgres.NewDb(ctx, &cfg.Postgres)
	if err != nil {
		return err
	}

	c, err := kafka.NewConsumer(kafka.Config{
		Brokers: []string{cfg.Kafka.Broker},
		Topic:   cfg.Kafka.ResultTopic,
	})
	if err != nil {
		return err
	}

	consumer := reader.New(c.Consumer, []string{cfg.Kafka.Broker}, cfg.Kafka.PredictTopic)
	predictionMessages := make(chan reader.PredictionMessage)

	err = consumer.Subscribe(ctx, predictionMessages)
	if err != nil {
		return err
	}

	// TODO: save images into s3 and results into postgresql
	// TODO: create api
	for {
		select {
		case msg := <-predictionMessages:

		}
	}

}
