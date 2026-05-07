package step_handler

import (
	"context"
	"fmt"

	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/proto/gen/go/sagaflow"
)

func (a *Adapter) Call(ctx context.Context, dto *domain.CallHandlerRequest) (*domain.CallHandlerResult, error) {
	client, found := a.clients[dto.Service]
	if !found {
		return &domain.CallHandlerResult{
			Status: domain.CallHandlerResultHandlerNotFound,
		}, nil
	}
	req := &sagaflow.HandleRequest{
		Meta: &sagaflow.StepExecutionMeta{
			SagaId:         dto.SagaInstanceId.String(),
			StepId:         dto.StepId,
			Action:         dto.Method,
			Attempt:        int32(dto.Attempt),
			IdempotencyKey: dto.IdempotencyKey,
		},
		Payload: dto.InputData.GetRaw(),
	}

	resp, err := client.Handle(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call failed: %w", err)
	}
	if !resp.Success {
		errText := "error"
		if resp.Error != nil {
			errText = *resp.Error
		}
		return &domain.CallHandlerResult{
			Status:    domain.CallHandlerResultUnprocessable,
			ErrorData: &errText,
		}, nil
	}

	return &domain.CallHandlerResult{
		Status: domain.CallHandlerResultSuccess,
	}, nil
}
