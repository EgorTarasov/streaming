package writer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EgorTarasov/streaming/framer/pkg/infrastructure/kafka"
	"github.com/IBM/sarama"
)

// FrameMessage формат сообщений логов
type FrameMessage struct {
	VideoId  int64  `json:"streamId"`
	FrameId  uint64 `json:"frameId"`
	RawFrame []byte
	Frame    string `json:"frame"`
}

// FrameKafkaSender структура с продюсером кафки
type FrameKafkaSender struct {
	producer *kafka.Producer
	topic    string
}

// NewFrameSender конструктор для отправки логов в kafka
// в сообщение содержит следующие поля (http запроса)
// msg - дополнительный текст для сообщения
// body - тело сообщения
// method - метод http запроса
// path - путь на который пришел запрос
// в качестве ключа для сообщений используется unix timestamp
func NewFrameSender(producer *kafka.Producer, topic string) *FrameKafkaSender {
	return &FrameKafkaSender{
		producer,
		topic,
	}
}

// SendMessage отправка полученного кадра
func (s *FrameKafkaSender) SendMessage(message FrameMessage) error {
	kafkaMsg, err := s.buildFrameMessage(message)
	if err != nil {
		fmt.Println("Send message marshal error", err)
		return err
	}

	s.producer.SendAsyncMessage(kafkaMsg)
	return nil
}

func (s *FrameKafkaSender) buildFrameMessage(message FrameMessage) (*sarama.ProducerMessage, error) {

	message.Frame = base64.StdEncoding.EncodeToString(message.RawFrame)

	msg, err := json.Marshal(message)
	curTime := time.Now().Unix()

	if err != nil {

		return nil, err
	}

	return &sarama.ProducerMessage{
		Topic:     s.topic,
		Value:     sarama.ByteEncoder(msg),
		Partition: -1,
		Key:       sarama.StringEncoder(fmt.Sprint(curTime)),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("framer-version"),
				Value: []byte("v0"),
			},
		},
	}, nil
}
