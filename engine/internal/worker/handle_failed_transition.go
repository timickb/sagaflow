package worker

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/utils"
)

// handleFailedTransition Обработать переход on.failed для шагов типа action/compensate
func (r *Runner) handleFailedTransition(
	event *broker.SagaStepResultEvent,
	sagaDef *domain.SagaDefinition,
	currentStepDef *domain.DefinitionStep,
	instance *domain.InstanceView,
	currentStep *domain.StepView,
) (result *eventHandleResult, err error) {
	var (
		now                  = time.Now()
		currentStepErrData   domain.InstanceContext
		instanceErrMessage   *string
		resultIsNotRetriable bool
	)
	if event.Error != nil {
		resultIsNotRetriable = !event.Error.Retriable
		instanceErrMessage = utils.Ptr(event.Error.String())
		currentStepErrData, err = domain.NewJsonInstanceContextFromAny(event.Error)
		if err != nil {
			return nil, fmt.Errorf("create instance context from error data struct: %w", err)
		}
	}

	if currentStepDef.Retry != nil && currentStepDef.Retry.MaxAttempts <= currentStep.Attempt && !resultIsNotRetriable {
		// Еще остались ретраи
		delay := currentStepDef.Retry.Delay.Milliseconds()
		if currentStepDef.Retry.Backoff == domain.RetryBackoffTypeExponential {
			delay *= 2 * int64(currentStep.Attempt)
		}
		nextExecution := currentStep.UpdatedAt.Add(time.Duration(delay) * time.Millisecond)

		return &eventHandleResult{
			InstanceTransitionDto: &domain.InstanceTransitionDto{
				Id:              instance.SagaId,
				ExecutionState:  utils.Ptr(domain.InstanceExecutionStateRunnable),
				NextExecutionAt: &nextExecution,
				ErrCode:         utils.Ptr(string(domain.InstanceErrorCodeHandler)),
				ErrMessage:      instanceErrMessage,
			},
		}, nil
	}

	// Если ретраи не предусмотрены или закончились - переходим по failed ветке
	nextStepName, ok := currentStepDef.Transitions[domain.OutcomeFailed]
	if !ok {
		// Перехода по failed нет -> завершить инстанс со статусом FAIL
		return NewEventHandleNoTerminalStateResult(instance.SagaId, currentStep.Name, now), nil
	}
	nextStepDef := utils.Find(sagaDef.Steps, func(s *domain.DefinitionStep) bool {
		return s.Id == nextStepName
	})
	if nextStepDef == nil {
		log.Error().Msgf("Next step %s not found for instance %v", event.Ref.StepName, instance.SagaId)
		return NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonStepNotFound), nil
	}

	return &eventHandleResult{
		InstanceTransitionDto: &domain.InstanceTransitionDto{
			Id:             instance.SagaId,
			NextStepName:   nextStepDef.Id,
			Status:         nextStepDef.Kind.ToInstanceStatus(instance.Status),
			ExecutionState: utils.Ptr(domain.InstanceExecutionStateRunnable),
		},
		StepCreateDto: &domain.StepCreateDto{
			InstanceId: instance.SagaId,
			StepName:   nextStepDef.Id,
			StepOrder:  currentStep.Order + 1,
			InputData:  instance.RuntimeContext.GetRaw(),
		},
		// ретраи исчерпаны -> исход текущего шага = failed
		StepUpdateDto: &domain.StepUpdateDto{
			InstanceId: instance.SagaId,
			StepName:   currentStep.Name,
			Status:     utils.Ptr(domain.StepStatusFailed),
			ErrorData:  currentStepErrData,
		},
	}, nil
}
