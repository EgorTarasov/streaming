package internal

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/EgorTarasov/streaming/framer/internal/config"
	"github.com/EgorTarasov/streaming/framer/internal/controller"
	"github.com/EgorTarasov/streaming/framer/internal/framer"
	internalkafka "github.com/EgorTarasov/streaming/framer/internal/kafka"
	"github.com/EgorTarasov/streaming/framer/internal/kafka/reader"
	"github.com/EgorTarasov/streaming/framer/internal/kafka/writer"
	"github.com/EgorTarasov/streaming/framer/internal/shared/commands"
	"github.com/EgorTarasov/streaming/framer/pkg/infrastructure/kafka"
	"github.com/EgorTarasov/streaming/framer/pkg/infrastructure/minios3"
	"github.com/IBM/sarama"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog/log"
)

type HandleFunc func(message *sarama.ConsumerMessage)

type Handler interface {
	Add(ctx context.Context, message reader.CommandMessage) error
	Remove(ctx context.Context, message reader.CommandMessage) error
	Status(ctx context.Context, message reader.CommandMessage) error
	Shutdown(ctx context.Context, message reader.CommandMessage) error
}

// const url = "rtsp://fake.kerberos.io/stream"

func Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg := config.MustNew("config.yaml")
	log.Info().Interface("minio cfg", cfg.S3)

	frameProcessor := framer.New()
	// test kafka config
	kafkaCfg := kafka.Config{
		Brokers: []string{cfg.Kafka.Broker},
		Topic:   cfg.Kafka.CmdTopic,
	}

	// create kafka topic
	err := internalkafka.Init(ctx, kafkaCfg.Brokers, cfg.Kafka.CmdTopic, cfg.Kafka.ResultTopic)
	if err != nil {

		return err
	}

	c, err := kafka.NewConsumer(kafkaCfg)
	if err != nil {

		return err
	}
	p, err := kafka.NewProducer(kafkaCfg.Brokers)
	if err != nil {
		return err
	}

	frameProducer := writer.NewFrameSender(p, cfg.Kafka.ResultTopic)
	responseProducer := writer.NewCommandSender(p, cfg.Kafka.ResponseTopic)

	s3 := minios3.MustNew(&cfg.S3)

	handler := controller.New(frameProcessor, responseProducer, frameProducer, &s3)

	consumer := reader.New(c.Consumer, kafkaCfg.Brokers, kafkaCfg.Topic)
	cmdMsg := make(chan reader.CommandMessage)

	err = consumer.Subscribe(ctx, cmdMsg)
	if err != nil {
		return err
	}

	for {
		select {
		case msg := <-cmdMsg:
			fmt.Println(msg)
			log.Info().Str("msg", msg.Command).Int64("videoId", msg.VideoId)
			switch msg.Command {

			case commands.Add:
				err = handler.Add(ctx, msg)
				if err != nil {
					log.Error().Str("err", err.Error()).Msg("err during commands.Add")
				}
			case commands.Status:
				err = handler.Status(ctx, msg)
				if err != nil {
					log.Error().Str("err", err.Error()).Msg("err during commands.Status")
				}
			case commands.Remove:

				err = handler.Remove(ctx, msg)
				if err != nil {
					log.Error().Str("err", err.Error()).Msg("err during commands.Status")
				}
			case commands.GetResultVideo:
				log.Info().Int64("video", msg.VideoId).Msg("processing result video")
				var filename string
				filename, err = handler.GetResultVideo(ctx, msg.VideoId)
				if err != nil {
					log.Error().Str("err", err.Error()).Msg("err during commands.GetResultVideo")
				} else {
					log.Info().Str("output_filename", filename).Msg("processed task")
				}

			case commands.Shutdown:
				err = handler.Shutdown(ctx, msg)
				if err != nil {
					log.Error().Str("err", err.Error()).Msg("err during commands.Shutdown")
				}

				cancel()
				return nil
			}
		}
	}
}

func TestS3(ctx context.Context) error {
	cfg := config.MustNew("config.yaml")
	s3 := minios3.MustNew(&cfg.S3)
	fmt.Println(cfg.S3)

	log.Info().Msg("test of s3 Storage")

	bucketName := "frames"
	objChan := s3.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Prefix: "9/"})
	tmpDir := "foobarTmp"
	tmpDirVideo := fmt.Sprintf("%s/%d", tmpDir, 9)
	err := os.Mkdir(tmpDirVideo, 0755)
	if err != nil {
		fmt.Println(err)
	}
	//defer os.RemoveAll(tmpDirVideo)

	for object := range objChan {
		obj, er := s3.GetObject(ctx, bucketName, object.Key, minio.GetObjectOptions{})
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

	cmd := exec.Command("ffmpeg", "-framerate", "30", "-pattern_type", "glob", "-i", fmt.Sprintf("%s/*.jpg", tmpDirVideo), "-c:v", "libx264", "-pix_fmt", "yuv420p", "out.mp4")
	_, err = cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
