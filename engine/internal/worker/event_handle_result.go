package worker

import (
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/utils"
)

type eventHandleResult struct {
	InstanceTransitionDto *domain.InstanceTransitionDto
	StepCreateDto         *domain.StepCreateDto
	StepUpdateDto         *domain.StepUpdateDto
}

func NewEventHandleFailedResult(instanceId uuid.UUID, reason domain.InstanceFailReason) *eventHandleResult {
	return &eventHandleResult{
		InstanceTransitionDto: &domain.InstanceTransitionDto{
			Id:      instanceId,
			Status:  utils.Ptr(domain.InstanceStatusFailed),
			ErrCode: utils.Ptr(string(reason)),
		},
	}
}

func NewEventHandleNoTerminalStateResult(instanceId uuid.UUID, currentStepName string, now time.Time) *eventHandleResult {
	return &eventHandleResult{
		InstanceTransitionDto: &domain.InstanceTransitionDto{
			Id:           instanceId,
			NextStepName: "",
			// Считаем результатом INCONSISTENT, если нет перехода в терминальный шаг
			Status:         utils.Ptr(domain.InstanceStatusInconsistent),
			ExecutionState: utils.Ptr(domain.InstanceExecutionStateRunnable),
		},
		StepUpdateDto: &domain.StepUpdateDto{
			InstanceId: instanceId,
			StepName:   currentStepName,
			Status:     utils.Ptr(domain.StepStatusCommitted),
		},
	}
}
