package service

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/EgorTarasov/streaming/api/internal/stubs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type orchestrator struct {
	host   string
	port   int
	client pb.VideoProcessingServiceClient
}

func New(host string, port int) *orchestrator {
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", host, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("err in init with %v %v", host, err))
	}
	client := pb.NewVideoProcessingServiceClient(conn)

	return &orchestrator{
		host:   host,
		port:   port,
		client: client,
	}
}

func (o *orchestrator) StartProcessingStream(ctx context.Context, rtspUrl string, title string) (int64, error) {
	resp, err := o.client.StartProcessing(ctx, &pb.StartProcessingRequest{
		VideoSource: &pb.StartProcessingRequest_Url{
			Url: rtspUrl,
		},
		Title: title,
	})
	if err != nil {
		return 0, err
	}
	return resp.JobId, nil
}

func (o *orchestrator) StartProcessingVideoFile(ctx context.Context, file []byte) (int64, error) {
	resp, err := o.client.StartProcessing(ctx, &pb.StartProcessingRequest{
		VideoSource: &pb.StartProcessingRequest_File{
			File: file,
		},
	})
	if err != nil {
		return 0, err
	}
	return resp.JobId, nil
}

type ProcessingStatus struct {
	Status          pb.Status
	Err             string
	ProcessedFrames int64
}

func (o *orchestrator) GetProcessingStatus(ctx context.Context, id int64) (ProcessingStatus, error) {
	resp, err := o.client.GetProcessingStatus(ctx, &pb.ProcessIdRequest{JobId: id})
	if err != nil {
		return ProcessingStatus{}, err
	}
	return ProcessingStatus{
		Status:          resp.Status,
		Err:             resp.ErrorMessage,
		ProcessedFrames: resp.Progress,
	}, nil
}

func (o *orchestrator) GetProcessingResult(ctx context.Context, id int64) (string, error) {
	resp, err := o.client.GetProcessingResult(ctx, &pb.ProcessResultRequest{
		JobId:      id,
		StartFrame: 0,
		EndFrame:   0,
	})
	if err != nil {
		return "", err
	}
	return resp.Result, err

}

func (o *orchestrator) StopProcessing(ctx context.Context, id int64) error {
	resp, err := o.client.StopProcessing(ctx, &pb.ProcessIdRequest{JobId: id})
	if err != nil {
		return err
	}
	if resp.Status == pb.Status_ERROR {
		return errors.New("error")
	}

	return nil
}
