package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/EgorTarasov/streaming/orchestrator/internal/controller/commands"
	"github.com/EgorTarasov/streaming/orchestrator/internal/repository/postgres"
	"github.com/EgorTarasov/streaming/orchestrator/internal/shared/servicies"
	"github.com/EgorTarasov/streaming/orchestrator/pkg/kafka"
	"github.com/EgorTarasov/streaming/orchestrator/pkg/minios3"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
)

type Controller struct {
	tasks    map[int64]context.CancelFunc
	mu       sync.RWMutex
	repo     *postgres.TaskRepo // TODO: move to interface
	consumer *kafka.Consumer
	producer *kafka.AsyncProducer
	s3       minios3.S3
}

func New(repo *postgres.TaskRepo, s3 minios3.S3, consumer *kafka.Consumer, producer *kafka.AsyncProducer) *Controller {
	return &Controller{
		tasks:    make(map[int64]context.CancelFunc),
		repo:     repo,
		s3:       s3,
		consumer: consumer,
		producer: producer,
	}
}

type FramerCommandMessage struct {
	RtspUrl string `json:"RtspUrl"`
	Command string `json:"Command"`
	VideoId int64  `json:"videoId"`
}

// AddTask adds task to the list of active tasks
func (c *Controller) AddTask(ctx context.Context, title, rtspUrl string) (int64, error) {
	id, err := c.repo.Create(ctx, title, rtspUrl)
	if err != nil {
		return 0, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	ctx, cancel := context.WithCancel(ctx)
	c.tasks[id] = cancel

	cmdMsg := FramerCommandMessage{
		RtspUrl: rtspUrl,
		Command: commands.FramerAddTask,
		VideoId: id,
	}
	jsonBytes, err := json.Marshal(cmdMsg)
	if err != nil {
		return 0, err
	}

	c.producer.SendAsyncMessage(ctx, &sarama.ProducerMessage{
		Topic: "frames",
		Key:   nil,
		Value: sarama.ByteEncoder(jsonBytes),
	})
	return id, nil
}

// {
//   "VideoId": 9,
//   "Metadata": {
//     "ul": "http://127.0.0.1:9000/videos/video_9.mp4%20?X-Amz-Algorithm\u003dAWS4-HMAC-SHA256\u0026X-Amz-Credential\u003dYv8wguEkKAqeyNEsbXpH%2F20240527%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date\u003d20240527T064823Z\u0026X-Amz-Expires\u003d172800\u0026X-Amz-SignedHeaders\u003dhost\u0026X-Amz-Signature\u003dd6f070fe818e106c8abf6371b3f8b29fc11082ec0d991581a450c9647f114313"
//   }
// }

// RemoveTask removes task from the list of active tasks
func (c *Controller) RemoveTask(ctx context.Context, id int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cancel, ok := c.tasks[id]
	if !ok {
		return nil
	}
	cmdMsg := FramerCommandMessage{
		RtspUrl: "",
		Command: commands.FramerRemoveTask,
		VideoId: id,
	}
	jsonBytes, err := json.Marshal(cmdMsg)
	if err != nil {
		return err
	}
	log.Info().Str("msg", string(jsonBytes)).Msg("remove task")
	c.producer.SendAsyncMessage(ctx, &sarama.ProducerMessage{
		Topic: "frames",
		Key:   nil,
		Value: sarama.ByteEncoder(jsonBytes),
	})
	c.producer.SendAsyncMessage(ctx, &sarama.ProducerMessage{
		Topic: "cmd-prediction",
		Key:   nil,
		Value: sarama.ByteEncoder(jsonBytes),
	})

	err = c.repo.UpdateStatus(ctx, id, "DONE")
	if err != nil {
		return err
	}
	cancel()
	delete(c.tasks, id)
	return nil
}

// ProcessResultTopic consumes messages from the responses topic
// use in separate goroutine
func (c *Controller) ProcessResultTopic(ctx context.Context) error {
	messageChan := make(chan *sarama.ConsumerMessage)
	err := c.consumer.Subscribe(ctx, "responses", messageChan)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case message := <-messageChan:
			serviceType := getServiceType(message.Headers)
			switch serviceType {
			case servicies.FramerServiceName:
				c.processFramer(ctx, message)
			case servicies.PredictionServiceName:
				c.processPrediction(ctx, message)
			}

		}

	}
}

type TaskStatusDto struct {
	VideoId         int64
	SplitFrames     int64
	ProcessedFrames int64
	Status          string
	CreatedAt       time.Time
}

// GetTaskStatus получение статуса задачи (кол-во нарезанных и обработанных кадров)
func (c *Controller) GetTaskStatus(ctx context.Context, taskId int64) (TaskStatusDto, error) {

	res, err := c.repo.GetTaskStatus(ctx, taskId)
	if err != nil {
		return TaskStatusDto{}, err
	}
	return TaskStatusDto{
		VideoId:         res.VideoId,
		SplitFrames:     res.SplitFrames,
		ProcessedFrames: res.ProcessedFrames,
		Status:          res.Status,
		CreatedAt:       res.CreatedAt,
	}, nil
}

// GetProcessingResult
func (c *Controller) GetProcessingResult(ctx context.Context, taskId int64) (string, error) {
	// sends message to kafka with framer if video not in db otherWise
	url, err := c.repo.GetResultVideo(ctx, taskId)
	if err != nil {
		cmdMsg := FramerCommandMessage{
			RtspUrl: "",
			Command: commands.GetResultVideoTask,
			VideoId: taskId,
		}
		jsonBytes, err := json.Marshal(cmdMsg)
		if err != nil {
			return "", err
		}
		c.producer.SendAsyncMessage(ctx, &sarama.ProducerMessage{
			Topic: "frames",
			Key:   nil,
			Value: sarama.ByteEncoder(jsonBytes),
		})
		return "", fmt.Errorf("video not processed")
	}
	return url, nil
}

func getServiceType(headers []*sarama.RecordHeader) string {
	for _, header := range headers {
		//log.Info().Str("header key", string(header.Key)).Str("value", string(header.Value)).Msg("decoding headers")
		if string(header.Key) == "service" {
			return string(header.Value)
		}
	}
	return ""

}

func getMessageType(headers []*sarama.RecordHeader) string {
	for _, header := range headers {
		if string(header.Key) == "type" {
			return string(header.Value)
		}
	}
	return ""
}
