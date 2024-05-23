package controller

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/EgorTarasov/streaming/responser/internal/kafka/reader"
	"github.com/EgorTarasov/streaming/responser/internal/repository/postgres"
	s3 "github.com/EgorTarasov/streaming/responser/pkg/infrastructure/minios3"
	"github.com/minio/minio-go/v7"
)

type Controller struct {
	s3         *s3.S3
	repo       *postgres.PredictionRepository
	bucketName string
}

func NewController(s3 *s3.S3, bucketName string, repo *postgres.PredictionRepository) *Controller {
	return &Controller{s3: s3, bucketName: bucketName, repo: repo}
}

func (c *Controller) SaveResult(ctx context.Context, msg reader.PredictionMessage) error {
	//	 save frame into s3, get s3 key and save into repo
	// decode frame from base64 into bytes and save as jpg

	frameBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(msg.Frame, "data:image/jpeg;base64,"))
	if err != nil {
		return err
	}
	s3Key := fmt.Sprintf("%d/%d.jpg", msg.VideoId, msg.FrameId)
	_, err = c.s3.PutObject(ctx, c.bucketName, s3Key, bytes.NewReader(frameBytes), int64(len(frameBytes)), minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	_, err = c.repo.SavePredictionResult(ctx, msg.VideoId, msg.FrameId, msg.Results, s3Key)
	if err != nil {
		return err
	}

	return nil
}
