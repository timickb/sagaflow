package controller

import (
	"context"
	"fmt"

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
	if err := c.svc.RebuildOrderProjection(ctx, req.Payload); err != nil {
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
