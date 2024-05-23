package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
)

type Consumer struct {
	brokers        []Broker
	SingleConsumer sarama.Consumer
}

func MustNewConsumer(config *Config, newest bool, commitInterval time.Duration) *Consumer {
	cfg := sarama.NewConfig()
	cfg.Consumer.Return.Errors = false
	cfg.Consumer.Offsets.AutoCommit.Enable = true
	cfg.Consumer.Offsets.AutoCommit.Interval = commitInterval

	if newest {
		cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	} else {
		cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	brokers := make([]string, len(config.Brokers))
	for idx, broker := range config.Brokers {
		brokers[idx] = fmt.Sprintf("%s:%d", broker.Host, broker.Port)
	}

	consumer, err := sarama.NewConsumer(brokers, cfg)
	if err != nil {
		log.Panic().Err(err)
	}

	return &Consumer{
		brokers:        config.Brokers,
		SingleConsumer: consumer,
	}
}

func (consumer *Consumer) Subscribe(ctx context.Context, topic string, out chan<- *sarama.ConsumerMessage) error {
	partitions, err := consumer.SingleConsumer.Partitions(topic)
	if err != nil {
		log.Err(err).Msg("failed to get partitions")
		return err
	}

	initialOffset := sarama.OffsetNewest

	for _, partition := range partitions {
		pc, err := consumer.SingleConsumer.ConsumePartition(topic, partition, initialOffset)
		if err != nil {
			log.Err(err).Int32("partition", partition).Msg("error consuming partition")
			return err
		}

		go consume(ctx, pc, partition, out)
	}
	return nil
}

func consume(ctx context.Context, pc sarama.PartitionConsumer, partition int32, out chan<- *sarama.ConsumerMessage) {
	log.Info().Int32("partition", partition).Msg("consumer started")
	for {
		select {
		case <-ctx.Done():
			if err := pc.Close(); err != nil {
				log.Err(err).Int32("partition", partition).Msg("failed to close consumer")
				return
			}
			log.Info().Int32("partition", partition).Msg("consumer closed")
			return

		case message := <-pc.Messages():
			out <- message
			//log.Info().
			//	Str("topic", message.Topic).
			//	Int32("partition", message.Partition).
			//	Str("key", string(message.Key)).
			//	Interface("headers", message.Headers).
			//	Interface("msg", message.Value).
			//	Int64("offset", message.Offset).
			//	Msg("kafka message claimed")
		}
	}
}
