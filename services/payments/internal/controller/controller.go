package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	sagaflow "github.com/timickb/sagaflow/proto/gen/go/sagaflow"
	"github.com/timickb/sagaflow/services/payments/internal/service"
)

type StepHandlerController struct {
	sagaflow.UnimplementedStepHandlerServiceServer
	svc *service.PaymentsService
}

func NewStepHandlerController(svc *service.PaymentsService) *StepHandlerController {
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
	case "Capture":
		return c.handleCapture(ctx, req)
	default:
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("unknown action: %s", action)),
		}, nil
	}
}

func (c *StepHandlerController) handleCapture(ctx context.Context, req *sagaflow.HandleRequest) (*sagaflow.HandleResponse, error) {
	var payload service.CapturePayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("parse payload: %v", err)),
		}, nil
	}

	orderID, err := uuid.Parse(payload.OrderID)
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("parse order_id: %v", err)),
		}, nil
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("parse user_id: %v", err)),
		}, nil
	}

	sagaID, err := uuid.Parse(req.Meta.GetSagaId())
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("parse saga_id: %v", err)),
		}, nil
	}

	_, err = c.svc.Capture(ctx, orderID, sagaID, req.Meta.GetStepId(), payload.Items, userID, payload.Amount)
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("capture failed: %v", err)),
		}, nil
	}

	return &sagaflow.HandleResponse{Success: true}, nil
}

func ptrString(s string) *string {
	return &s
}

// Ensure StepHandlerController implements StepHandlerServiceServer
var _ sagaflow.StepHandlerServiceServer = (*StepHandlerController)(nil)
