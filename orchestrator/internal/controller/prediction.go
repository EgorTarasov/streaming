package controller

import (
	"context"
	"encoding/json"

	"github.com/EgorTarasov/streaming/orchestrator/internal/controller/commands"
	"github.com/EgorTarasov/streaming/orchestrator/internal/controller/messages"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
)

type PredictionHealthCheckMessage struct {
	VideoId  int64 `json:"VideoId"`
	Metadata struct {
		Frames int64 `json:"Frames"`
	} `json:"Metadata"`
}

type PredictionRemoveMessage struct {
	VideoId int64  `json:"VideoId"`
	Status  string `json:"Status"`
}

func (c *Controller) processPrediction(ctx context.Context, message *sarama.ConsumerMessage) {
	msgType := getMessageType(message.Headers)
	switch msgType {
	case messages.HealthCheck:
		log.Info().Msg("prediction healthcheck")
		var healthCheck PredictionHealthCheckMessage
		if err := json.Unmarshal(message.Value, &healthCheck); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal prediction healthcheck message")
			return
		}
		if err := c.repo.UpdatePredictedFrames(ctx, healthCheck.VideoId, healthCheck.Metadata.Frames); err != nil {
			log.Error().Err(err).Msg("failed to update predicted frames")
			return
		}
	case commands.PredictionRemoveTask:
		log.Info().Msg("prediction remove")
		var remove PredictionRemoveMessage
		if err := json.Unmarshal(message.Value, &remove); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal prediction remove message")
			return
		}
		if err := c.repo.UpdateStatus(ctx, remove.VideoId, remove.Status); err != nil {
			log.Error().Err(err).Msg("failed to update prediction status")
			return
		}
	}

	// TODO: обработать сообщения от prediction (два типа сообщений)
}
