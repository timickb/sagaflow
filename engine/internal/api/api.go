package api

import (
	"github.com/timickb/sagaflow/engine/internal/domain"
	pb "github.com/timickb/sagaflow/proto/gen/go/sagaflow"
)

type SagaflowServer struct {
	pb.UnimplementedSagaflowServiceServer

	instanceUsecase domain.InstanceUsecase
}

// NewSagaflowServer creates a new SagaflowServer instance.
func NewSagaflowServer(instanceUsecase domain.InstanceUsecase) *SagaflowServer {
	return &SagaflowServer{
		instanceUsecase: instanceUsecase,
	}
}
