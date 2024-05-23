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
func (s *server) GetProcessingStatus(ctx context.Context, in *pb.ProcessIdRequest) (*pb.GetProcessingStatusResponse, error) {

	panic("implement me")
}
func (s *server) GetProcessingResult(ctx context.Context, in *pb.ProcessResultRequest) (*pb.GetProcessingResultResponse, error) {
	panic("implement me")
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
