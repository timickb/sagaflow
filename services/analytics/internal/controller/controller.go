package controller

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	sagaflow "github.com/timickb/sagaflow/proto/gen/go/sagaflow"
	"github.com/timickb/sagaflow/services/analytics/internal/service"
)

type StepHandlerController struct {
	sagaflow.UnimplementedStepHandlerServiceServer
	svc *service.AnalyticsService
}

func NewStepHandlerController(svc *service.AnalyticsService) *StepHandlerController {
	return &StepHandlerController{svc: svc}
}

// Handle обрабатывает запросы саги
func (c *StepHandlerController) Handle(ctx context.Context, req *sagaflow.HandleRequest) (*sagaflow.HandleResponse, error) {
	if req.Meta == nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString("meta is required"),
		}, nil
	}

	action := req.Meta.GetAction()
	switch action {
	case "RebuildOrderProjection":
		return c.handleRebuildOrderProjection(ctx, req)
	default:
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("unknown action: %s", action)),
		}, nil
	}
}

func (c *StepHandlerController) handleRebuildOrderProjection(ctx context.Context, req *sagaflow.HandleRequest) (*sagaflow.HandleResponse, error) {
	sagaId, err := uuid.Parse(req.Meta.GetSagaId())
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("invalid saga_id: %v", err)),
		}, nil
	}

	meta := service.RebuildOrderProjectionMeta{
		SagaId:   sagaId,
		StepName: req.Meta.GetStepId(),
	}

	if err = c.svc.RebuildOrderProjection(ctx, req.Payload, meta); err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("rebuild projection failed: %v", err)),
		}, nil
	}

	return &sagaflow.HandleResponse{Success: true}, nil
}

func ptrString(s string) *string {
	return &s
}

// Ensure StepHandlerController implements StepHandlerServiceServer
var _ sagaflow.StepHandlerServiceServer = (*StepHandlerController)(nil)
