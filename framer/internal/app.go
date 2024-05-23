package internal

import (
	"context"
	"fmt"

	"github.com/EgorTarasov/streaming/framer/internal/config"
	"github.com/EgorTarasov/streaming/framer/internal/controller"
	"github.com/EgorTarasov/streaming/framer/internal/framer"
	internalkafka "github.com/EgorTarasov/streaming/framer/internal/kafka"
	"github.com/EgorTarasov/streaming/framer/internal/kafka/reader"
	"github.com/EgorTarasov/streaming/framer/internal/kafka/writer"
	"github.com/EgorTarasov/streaming/framer/internal/shared/commands"
	"github.com/EgorTarasov/streaming/framer/pkg/infrastructure/kafka"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
)

type HandleFunc func(message *sarama.ConsumerMessage)

type Handler interface {
	Add(ctx context.Context, message reader.CommandMessage) error
	Remove(ctx context.Context, message reader.CommandMessage) error
	Status(ctx context.Context, message reader.CommandMessage) error
	Shutdown(ctx context.Context, message reader.CommandMessage) error
}

// const url = "rtsp://fake.kerberos.io/stream"

func Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg := config.MustNew()

	frameProcessor := framer.New()
	// test kafka config
	kafkaCfg := kafka.Config{
		Brokers: []string{cfg.Kafka.Broker},
		Topic:   cfg.Kafka.CmdTopic,
	}

	// create kafka topic
	err := internalkafka.Init(ctx, kafkaCfg.Brokers, cfg.Kafka.CmdTopic, cfg.Kafka.ResultTopic)
	if err != nil {

		return err
	}

	c, err := kafka.NewConsumer(kafkaCfg)
	if err != nil {

		return err
	}
	p, err := kafka.NewProducer(kafkaCfg.Brokers)
	if err != nil {
		return err
	}

	frameProducer := writer.NewFrameSender(p, cfg.Kafka.ResultTopic)
	responseProducer := writer.NewCommandSender(p, cfg.Kafka.ResponseTopic)

	handler := controller.New(frameProcessor, responseProducer, frameProducer)

	consumer := reader.New(c.Consumer, kafkaCfg.Brokers, kafkaCfg.Topic)
	cmdMsg := make(chan reader.CommandMessage)

	err = consumer.Subscribe(ctx, cmdMsg)
	if err != nil {
		return err
	}

	for {
		select {
		case msg := <-cmdMsg:
			fmt.Println(msg)
			log.Info().Str("msg", msg.Command).Int64("videoId", msg.VideoId)
			switch msg.Command {

			case commands.Add:
				err = handler.Add(ctx, msg)
				if err != nil {
					log.Error().Str("err", err.Error()).Msg("err during commands.Add")
				}
			case commands.Status:
				err = handler.Status(ctx, msg)
				if err != nil {
					log.Error().Str("err", err.Error()).Msg("err during commands.Status")
				}
			case commands.Remove:

				err = handler.Remove(ctx, msg)
				if err != nil {
					log.Error().Str("err", err.Error()).Msg("err during commands.Status")
				}
			case commands.Shutdown:
				err = handler.Shutdown(ctx, msg)
				if err != nil {
					log.Error().Str("err", err.Error()).Msg("err during commands.Shutdown")
				}

				cancel()
				return nil
			}
		}
	}
}
