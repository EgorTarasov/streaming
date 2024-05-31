package server

import (
	"context"

	"github.com/EgorTarasov/streaming/orchestrator/internal/controller"
	"github.com/EgorTarasov/streaming/orchestrator/internal/grpc/pb"
)

type server struct {
	ctrl *controller.Controller
	pb.UnimplementedVideoProcessingServiceServer
}

func New(ctrl *controller.Controller) *server {
	return &server{ctrl: ctrl}
}

func (s *server) StartProcessing(ctx context.Context, in *pb.StartProcessingRequest) (*pb.StartProcessingResponse, error) {
	newId, err := s.ctrl.AddTask(ctx, in.Title, in.GetUrl())
	if err != nil {
		return nil, err
	}
	return &pb.StartProcessingResponse{
		JobId: newId,
	}, nil
}

func mapStatus(status string) pb.Status {
	switch status {
	case "PENDING":
		return pb.Status_PENDING
	case "PROCESSING":
		return pb.Status_PROCESSING
	case "DONE":
		return pb.Status_DONE
	case "ERROR":
		return pb.Status_ERROR
	}
	return pb.Status_ERROR
}

func (s *server) GetProcessingStatus(ctx context.Context, in *pb.ProcessIdRequest) (*pb.GetProcessingStatusResponse, error) {
	res, err := s.ctrl.GetTaskStatus(ctx, in.JobId)
	if err != nil {
		return nil, err
	}
	return &pb.GetProcessingStatusResponse{
		Status:       mapStatus(res.Status),
		ErrorMessage: "",
		Progress:     res.ProcessedFrames,
		SplitFrames:  res.SplitFrames,
	}, nil
}
func (s *server) GetProcessingResult(ctx context.Context, in *pb.ProcessResultRequest) (*pb.GetProcessingResultResponse, error) {
	url, err := s.ctrl.GetProcessingResult(ctx, in.JobId)
	if err != nil {
		return nil, err
	}
	return &pb.GetProcessingResultResponse{
		Result: url,
	}, nil
}
func (s *server) StopProcessing(ctx context.Context, in *pb.ProcessIdRequest) (*pb.StopProcessingResponse, error) {
	if err := s.ctrl.RemoveTask(ctx, in.JobId); err != nil {
		return nil, err
	}
	return &pb.StopProcessingResponse{
		Status:       pb.Status_DONE,
		ErrorMessage: "",
	}, nil
}
