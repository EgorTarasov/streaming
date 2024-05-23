package reader

import (
	"encoding/json"

	"context"

	"github.com/IBM/sarama"
)

type consumer struct {
	brokers []string
	topic   string
	sarama.Consumer
}

type PredictMessageHandler = func(ctx context.Context, message PredictionMessage)

func New(c sarama.Consumer, brokers []string, topic string) *consumer {
	return &consumer{
		brokers:  brokers,
		topic:    topic,
		Consumer: c,
	}
}

type PredictionMessage struct {
	VideoId int64       `json:"VideoId"`
	FrameId int64       `json:"FrameId"`
	Frame   string      `json:"Frame"`
	Results interface{} `json:"Results"`
}

// Subscribe Запускает обработку handler по заданным топикам
func (c *consumer) Subscribe(_ context.Context, ch chan<- PredictionMessage) error {

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
			var msg PredictionMessage
			for message := range pc.Messages() {
				_ = json.Unmarshal(message.Value, &msg)
				ch <- msg
			}
		}(pc, partition)
	}

	return nil
}
