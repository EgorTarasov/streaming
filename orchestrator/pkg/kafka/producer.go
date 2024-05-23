package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
)

type AsyncProducer struct {
	brokers  []Broker
	producer sarama.AsyncProducer
}

func MustNewAsyncProducer(config *Config, partitioner sarama.PartitionerConstructor, acks sarama.RequiredAcks) *AsyncProducer {
	cfg := sarama.NewConfig()

	cfg.Producer.Partitioner = partitioner
	cfg.Producer.RequiredAcks = acks

	cfg.Producer.Return.Successes = false
	cfg.Producer.Return.Errors = true

	brokers := make([]string, len(config.Brokers))
	for idx, broker := range config.Brokers {
		brokers[idx] = fmt.Sprintf("%s:%d", broker.Host, broker.Port)
	}

	asyncProducer, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		log.Panic().Err(err)
	}

	go func() {
		for e := range asyncProducer.Errors() {
			log.Err(e).Msg("kafka.AsyncProducer")
		}
	}()

	return &AsyncProducer{
		brokers:  config.Brokers,
		producer: asyncProducer,
	}
}

func (k *AsyncProducer) SendAsyncMessage(ctx context.Context, message *sarama.ProducerMessage) {
	k.producer.Input() <- message
}

func (k *AsyncProducer) Close() {
	if err := k.producer.Close(); err != nil {
		log.Err(err).Msg("kafka.AsyncProducer.Close")
	}
}
