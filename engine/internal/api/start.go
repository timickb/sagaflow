package api

import (
	"context"

	pb "github.com/timickb/sagaflow/proto/gen/go/sagaflow"
	"google.golang.org/grpc/codes"
)

// Start - запустить новый экземпляр саги
func (s *SagaflowServer) Start(ctx context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	dto, err := startRequestToDto(req)
	if err != nil {
		return &pb.StartResponse{
			Status: errorToStatus(err),
		}, nil
	}

	instanceId, err := s.instanceUsecase.Start(ctx, dto)
	if err != nil {
		return &pb.StartResponse{
			Status: errorToStatus(err),
		}, nil
	}

	return &pb.StartResponse{
		Status:         &pb.Status{Code: int32(codes.OK)},
		SagaInstanceId: instanceId.String(),
	}, nil
}
