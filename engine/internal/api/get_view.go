package api

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/timickb/sagaflow/proto/gen/go/sagaflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetView - получить представление экземпляра саги
func (s *SagaflowServer) GetView(ctx context.Context, req *pb.GetViewRequest) (*pb.GetViewResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	instanceId, err := uuid.Parse(req.SagaInstanceId)
	if err != nil {
		return &pb.GetViewResponse{
			Status: &pb.Status{
				Code:  int32(codes.InvalidArgument),
				Error: &pb.Error{Message: "invalid saga instance id"},
			},
		}, nil
	}

	view, err := s.instanceUsecase.GetView(ctx, instanceId)
	if err != nil {
		return &pb.GetViewResponse{
			Status: errorToStatus(err),
		}, nil
	}

	return &pb.GetViewResponse{
		Status: &pb.Status{Code: int32(codes.OK)},
		View:   instanceViewToProto(view),
	}, nil
}
