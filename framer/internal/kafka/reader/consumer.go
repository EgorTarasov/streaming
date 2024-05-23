package reader

import (
	"encoding/json"

	"github.com/IBM/sarama"
)

import (
	"context"
)

type consumer struct {
	brokers []string
	topic   string
	sarama.Consumer
}

type CommandMessageHandler = func(ctx context.Context, message CommandMessage)

func New(c sarama.Consumer, brokers []string, topic string) *consumer {
	return &consumer{
		brokers:  brokers,
		topic:    topic,
		Consumer: c,
	}
}

type CommandMessage struct {
	RtspUrl string `json:"RtspUrl"`
	Command string `json:"Command"`
	VideoId int64  `json:"videoId"`
}

// Subscribe Запускает обработку handler по заданным топикам
func (c *consumer) Subscribe(_ context.Context, ch chan<- CommandMessage) error {

	// получаем все партиции топика
	partitionList, err := c.Consumer.Partitions(c.topic)

	if err != nil {
		return err
	}

	initialOffset := sarama.OffsetNewest

	for _, partition := range partitionList {
		pc, err := c.Consumer.ConsumePartition(c.topic, partition, initialOffset)

		if err != nil {
			return err
		}

		go func(pc sarama.PartitionConsumer, partition int32) {
			var msg CommandMessage
			for message := range pc.Messages() {
				_ = json.Unmarshal(message.Value, &msg)
				ch <- msg
			}
		}(pc, partition)
	}

	return nil
}
