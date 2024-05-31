package controller

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/EgorTarasov/streaming/framer/internal/framer"
	"github.com/EgorTarasov/streaming/framer/internal/kafka/reader"
	"github.com/EgorTarasov/streaming/framer/internal/kafka/writer"
	"github.com/EgorTarasov/streaming/framer/internal/shared/commands"
	"github.com/EgorTarasov/streaming/framer/internal/shared/status"
	"github.com/EgorTarasov/streaming/framer/pkg/infrastructure/minios3"
	"github.com/minio/minio-go/v7"
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
	s3               *minios3.S3
}

func New(f Framer, rp ResponseProducer, fp FrameProducer, s3 *minios3.S3) *controller {
	return &controller{f, rp, fp, s3}
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

// GetResultVideo собирает видео из обработанных кадров из s3 хранилища
// taskId - задача обработки видео потока кадры которого будут составлять итоговое видео
// возвращает ссылку на видео в s3
func (c *controller) GetResultVideo(ctx context.Context, taskId int64) (string, error) {

	bucketName := "frames"

	// получение кадров из s3 формат хранения <task_id>/<frame_id>.jpg
	objChan := c.s3.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Prefix: fmt.Sprintf("%d/", taskId)})
	tmpDir := "foobarTmp"
	_ = os.Mkdir(tmpDir, 0755)

	tmpDirVideo := fmt.Sprintf("%s/%d", tmpDir, 9)
	err := os.Mkdir(tmpDirVideo, 0755)
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		if er := os.RemoveAll(tmpDirVideo); er != nil {
			log.Err(er).Msg("error during deleting temp dir")
		}

	}()

	for object := range objChan {
		obj, er := c.s3.GetObject(ctx, bucketName, object.Key, minio.GetObjectOptions{})
		log.Info().Interface("obj", obj).Msg("new object from s3")
		if er != nil {
			log.Err(er).Msg("error during downloading file")
		}
		rawFile, er := io.ReadAll(obj)
		if er != nil {
			log.Err(er).Str("file_key", object.Key).Msg("err during file retrival")
		}
		er = os.WriteFile(fmt.Sprintf("%s/%s", tmpDir, object.Key), rawFile, 0755)
		if er != nil {
			fmt.Println(er)
			log.Err(er).Str("file_key", object.Key).Msg("err saving file")
		}
	}
	videoName := fmt.Sprintf("video_%d.mp4 ", taskId)

	//cmd := exec.Command("ffmpeg", "-framerate", "30", "-pattern_type", "glob", "-i", fmt.Sprintf("%s/*.jpg", tmpDirVideo), "-c:v", "libx264", "-pix_fmt", "yuv420p", videoName)
	cmd := exec.Command("ffmpeg", "-framerate", "30", "-pattern_type", "glob", "-i", fmt.Sprintf("%s/*.jpg", tmpDirVideo), "-c:v", "libx264", "-pix_fmt", "yuv420p", "out.mp4")
	defer os.Remove("out.mp4")
	fmt.Println(cmd)
	_, err = cmd.Output()
	if err != nil {
		log.Err(err).Msg("ffmpeg err")
		return "", err
	}
	videoBytes, err := os.ReadFile("out.mp4")
	if err != nil {
		return "", err
	}
	log.Info().Str("videoName", videoName).Msg("created name for new video")
	videoObject, err := c.s3.PutObject(ctx, "videos", videoName, bytes.NewReader(videoBytes), int64(len(videoBytes)), minio.PutObjectOptions{ContentType: "video/mp4"})
	if err != nil {
		return "", err
	}

	videoUrl, err := c.s3.PresignedGetObject(ctx, "videos", videoObject.Key, time.Duration(time.Hour*48), url.Values{})
	if err != nil {
		log.Err(err).Msg("err getting video url")
	}

	// TODO: retry for sending message
	err = c.responseProducer.SendMessage(writer.CommandResponseMessage{
		VideoId: taskId,
		Metadata: struct {
			Url string `json:"url"`
		}{
			Url: videoUrl.String(),
		},
	}, commands.GetResultVideo)
	if err != nil {
		log.Err(err).Msg("err sending response to response channel")
	}

	return videoUrl.String(), nil
}
