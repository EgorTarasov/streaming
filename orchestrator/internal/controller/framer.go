package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EgorTarasov/streaming/orchestrator/internal/controller/commands"
	"github.com/EgorTarasov/streaming/orchestrator/internal/controller/messages"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
)

type FramerHealthCheckMessage struct {
	VideoId  int64 `json:"VideoId"`
	Metadata struct {
		Status    string    `json:"Status"`
		Frames    int64     `json:"Frames"`
		TimeStamp time.Time `json:"TimeStamp"`
	} `json:"Metadata"`
}

type FramerAddedMessage struct {
	VideoId  int64 `json:"VideoId"`
	Metadata struct {
		Status    string    `json:"Status"`
		TimeStamp time.Time `json:"TimeStamp"`
	} `json:"Metadata"`
}

type FramerRemovedMessage struct {
	VideoId  int64 `json:"VideoId"`
	Metadata struct {
		Status    string    `json:"Status"`
		TimeStamp time.Time `json:"TimeStamp"`
	} `json:"Metadata"`
}

func (c *Controller) processFramer(ctx context.Context, message *sarama.ConsumerMessage) {
	// TODO: обработать сообщения от framer (три типа сообщений)
	msgType := getMessageType(message.Headers)
	fmt.Println(msgType)
	switch msgType {
	case messages.HealthCheck:
		log.Info().Msg("framer healthcheck")
		var healthCheck FramerHealthCheckMessage
		if err := json.Unmarshal(message.Value, &healthCheck); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal message")
			return
		}
		//log.Info().Interface("healthCheck", healthCheck).Msg("health check")
		if err := c.repo.UpdateSplitFrames(ctx, healthCheck.VideoId, healthCheck.Metadata.Status, healthCheck.Metadata.Frames); err != nil {
			log.Error().Err(err).Msg("failed to update status")
			return
		}
	case commands.FramerAddTask:
		var added FramerAddedMessage
		if err := json.Unmarshal(message.Value, &added); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal message")
			return
		}
		if err := c.repo.UpdateStatus(ctx, added.VideoId, added.Metadata.Status); err != nil {
			log.Error().Err(err).Msg("failed to update status")
			return
		}
	case commands.FramerRemoveTask:
		var removed FramerRemovedMessage
		if err := json.Unmarshal(message.Value, &removed); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal message")
			return
		}
		if err := c.repo.UpdateStatus(ctx, removed.VideoId, removed.Metadata.Status); err != nil {
			log.Error().Err(err).Msg("failed to update status")
			return
		}
	case commands.FramerStatusTask:
		var status FramerHealthCheckMessage
		if err := json.Unmarshal(message.Value, &status); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal message")
			return
		}
		if err := c.repo.UpdateStatus(ctx, status.VideoId, status.Metadata.Status); err != nil {
			log.Error().Err(err).Msg("failed to update status")
			return
		}
	case commands.GetResultVideoTask:
		var videoResponse messages.VideoResultMessage
		if err := json.Unmarshal(message.Value, &videoResponse); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal videoResponse message")
		}
		if id, err := c.repo.SaveResultVideo(ctx, videoResponse.VideoId, videoResponse.Metadata.Url); err != nil {
			log.Error().Err(err).Int64("taskId", id).Msg("failed to update status")
			return
		}

	default:
		var jsonValue interface{}
		if err := json.Unmarshal(message.Value, &jsonValue); err != nil {
			log.Error().Err(err).Msg("json decoding for unknown msgTypea")
		}

		log.Info().Interface("unknownJson:", jsonValue).Msg("unknown message type")
	}

}
