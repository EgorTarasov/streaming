package writer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/EgorTarasov/streaming/framer/internal/shared/commands"
	"github.com/EgorTarasov/streaming/framer/pkg/infrastructure/kafka"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
)

// CommandResponseMessage формат сообщений логов
type CommandResponseMessage struct {
	VideoId  int64       `json:"VideoId"`
	Metadata interface{} `json:"Metadata"`
}

// CommandResponseSender структура с продюсером кафки
type CommandResponseSender struct {
	producer *kafka.Producer
	topic    string
}

// NewCommandSender конструктор для отправки логов в kafka
// в сообщение содержит следующие поля (http запроса)
// msg - дополнительный текст для сообщения
// body - тело сообщения
// method - метод http запроса
// path - путь на который пришел запрос
// в качестве ключа для сообщений используется unix timestamp
func NewCommandSender(producer *kafka.Producer, topic string) *CommandResponseSender {
	return &CommandResponseSender{
		producer,
		topic,
	}
}

// SendMessage отправка полученного кадра
func (s *CommandResponseSender) SendMessage(message CommandResponseMessage, msgType string) error {
	if msgType != commands.StatusMsgType && msgType != commands.ResponseMsgType && msgType != commands.HealthCheck && msgType != commands.Add && msgType != commands.Remove {
		return fmt.Errorf("invalid message type")
	}

	kafkaMsg, err := s.buildMessage(message, msgType)
	if err != nil {
		fmt.Println("Send message marshal error", err)
		return err
	}
	s.producer.SendAsyncMessage(kafkaMsg)
	return nil
}

func (s *CommandResponseSender) buildMessage(message CommandResponseMessage, msgType string) (*sarama.ProducerMessage, error) {
	msg, err := json.Marshal(message)
	log.Info().Interface("sending", message).Bytes("result", msg).Msg("encoding kafka msg")

	curTime := time.Now().Unix()

	if err != nil {
		return nil, err
	}

	messageTypeHeader := sarama.RecordHeader{
		Key:   []byte("type"),
		Value: []byte(msgType),
	}
	log.Info().Interface("value", message).Msg("sending response msg")
	return &sarama.ProducerMessage{
		Topic:     s.topic,
		Value:     sarama.ByteEncoder(msg),
		Partition: -1,
		Key:       sarama.StringEncoder(fmt.Sprint(curTime)),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("service"),
				Value: []byte("framer"),
			}, messageTypeHeader,
		},
	}, nil
}
