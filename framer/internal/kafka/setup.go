package kafka

import (
	"context"
	"errors"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
)

// Init подготовка kafka к работе
// создание топиков и тд
func Init(ctx context.Context, brokers []string, consumerTopic string, writerTopic string) error {
	config := sarama.NewConfig()

	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return err
	}
	defer admin.Close()

	// создание топиков
	err = admin.CreateTopic(consumerTopic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
	if err != nil && !errors.Is(err, sarama.ErrTopicAlreadyExists) {

		return err

	}
	err = admin.CreateTopic(writerTopic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
	if err != nil && !errors.Is(err, sarama.ErrTopicAlreadyExists) {
		return err
	}
	log.Info().Msg("created required topics")

	return nil
}
