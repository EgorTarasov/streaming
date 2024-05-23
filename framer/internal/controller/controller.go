package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/EgorTarasov/streaming/framer/internal/framer"
	"github.com/EgorTarasov/streaming/framer/internal/kafka/reader"
	"github.com/EgorTarasov/streaming/framer/internal/kafka/writer"
	"github.com/EgorTarasov/streaming/framer/internal/shared/commands"
	"github.com/EgorTarasov/streaming/framer/internal/shared/status"
	"github.com/rs/zerolog/log"
)

type Framer interface {
	AddStream(ctx context.Context, videoId int64, rtspUrl string, processFunc framer.ProcessFrameFunc) error
	RemoveStream(_ context.Context, videoId int64) error
	StreamStatus(_ context.Context, videoId int64) (framer.TaskInfo, error)
	Status(_ context.Context) ([]framer.TaskInfo, error)
}

type ResponseProducer interface {
	SendMessage(message writer.CommandResponseMessage, msgType string) error
}

type FrameProducer interface {
	SendMessage(message writer.FrameMessage) error
}

type controller struct {
	f                Framer
	responseProducer ResponseProducer
	fp               FrameProducer
}

func New(f Framer, rp ResponseProducer, fp FrameProducer) *controller {
	return &controller{f, rp, fp}
}

func (c *controller) processFrame(ctx context.Context, frameId int64, videoId int64, frame []byte) {

	//log.Info().Int("frame len", len(frame)).Msg("new frame")
	er := c.fp.SendMessage(writer.FrameMessage{
		VideoId:  videoId,
		FrameId:  uint64(frameId),
		RawFrame: frame,
	})
	if er != nil {
		log.Error().Str("err", er.Error()).Msg("error during sending frame")
	}
	if frameId%10 == 0 {
		//log.Info().Int64("frameId", frameId).Msg("frameId")
		// отправляем статус обработки (каждый 10 кадр) health check
		er = c.responseProducer.SendMessage(writer.CommandResponseMessage{
			VideoId: videoId,
			Metadata: struct {
				Status    string    `json:"Status"`
				Frames    int64     `json:"Frames"`
				TimeStamp time.Time `json:"TimeStamp"`
			}{
				Status:    status.PROCESSING,
				TimeStamp: time.Now(),
				Frames:    frameId,
			},
		}, commands.HealthCheck)
		if er != nil {
			log.Err(er).Msg("error during sending response")
		}
	}

}

// Add добавляет видеопоток в обработку
func (c *controller) Add(ctx context.Context, message reader.CommandMessage) error {
	er := c.f.AddStream(ctx, message.VideoId, message.RtspUrl, c.processFrame)
	if er != nil {
		return er
		//log.Error().Str("err", er.Error()).Msg("err during commands.Add")
	}
	er = c.responseProducer.SendMessage(writer.CommandResponseMessage{
		VideoId: message.VideoId,
		Metadata: struct {
			Status    string    `json:"Status"`
			TimeStamp time.Time `json:"TimeStamp"`
		}{
			Status:    status.PENDING,
			TimeStamp: time.Now(),
		},
	}, commands.Add) // +
	if er != nil {
		return er
		//log.Error().Str("err", er.Error()).Msg("err during commands.Add")
	}
	return nil
}

// Remove удаление потока из обработки
func (c *controller) Remove(ctx context.Context, message reader.CommandMessage) error {
	er := c.f.RemoveStream(ctx, message.VideoId)
	if er != nil {
		return er
	}
	er = c.responseProducer.SendMessage(writer.CommandResponseMessage{
		VideoId: message.VideoId,
		Metadata: struct {
			Status    string    `json:"Status"`
			TimeStamp time.Time `json:"TimeStamp"`
		}{
			Status:    status.DONE,
			TimeStamp: time.Now(),
		},
	}, commands.Remove) // +
	if er != nil {
		// log.Error().Str("err", er.Error()).Msg("err during commands.Add")
		return er

	}
	return nil
}

// Status получение статуса обработки потока
func (c *controller) Status(ctx context.Context, message reader.CommandMessage) error {
	state, er := c.f.Status(ctx)
	fmt.Println(state)
	if er != nil {
		return er
		//log.Error().Str("err", er.Error()).Msg("err during commands.Status")
	}
	er = c.responseProducer.SendMessage(writer.CommandResponseMessage{
		VideoId: message.VideoId,
		Metadata: struct {
			Status    string            `json:"Status"`
			TimeStamp time.Time         `json:"TimeStamp"`
			CmdStatus []framer.TaskInfo `json:"ServiceStatus"`
		}{
			Status:    status.PROCESSING,
			TimeStamp: time.Now(),
			CmdStatus: state,
		},
	}, commands.Status) // +
	if er != nil {
		return er
	}
	return nil
}

// Shutdown завершение выполнения программы
func (c *controller) Shutdown(_ context.Context, message reader.CommandMessage) error {
	err := c.responseProducer.SendMessage(writer.CommandResponseMessage{
		VideoId: message.VideoId,
		Metadata: struct {
			Status    string    `json:"Status"`
			TimeStamp time.Time `json:"TimeStamp"`
		}{
			Status:    status.DONE,
			TimeStamp: time.Now(),
		},
	}, commands.Shutdown)
	return err
}
