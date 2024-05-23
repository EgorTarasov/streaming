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

		log.Info().Str("topic", consumerTopic).Msg("topic  already exists")

	}
	// FIXME: move init into orchestrator
	err = admin.CreateTopic("cmd-prediction", &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
	if err != nil && !errors.Is(err, sarama.ErrTopicAlreadyExists) {

		log.Info().Str("topic", "cmd-prediction").Msg("topic  already exists")

	}

	err = admin.CreateTopic(writerTopic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
	if err != nil && !errors.Is(err, sarama.ErrTopicAlreadyExists) {
		log.Info().Msg("topic  already exists")
	}
	log.Info().Str("topic", writerTopic).Msg("topic  already exists")

	return nil
}
