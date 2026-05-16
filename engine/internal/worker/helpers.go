package worker

import (
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/utils"
)

func calculateNextRetry(retryPolicy *domain.RetryPolicy, currentStep *domain.StepView) time.Time {
	delay := retryPolicy.Delay.Milliseconds()
	if retryPolicy.Backoff == domain.RetryBackoffTypeExponential {
		delay *= 2 * int64(currentStep.Attempt)
	}
	return currentStep.UpdatedAt.Add(time.Duration(delay) * time.Millisecond)
}

// mergeStepOutputToContext - добавить в runtime контекст сценария данные, объявленные
// в параметре output выполненного шага
func mergeStepOutputToContext(
	instanceCtx domain.InstanceContext,
	currentStepDef *domain.DefinitionStep,
	event *broker.SagaStepResultEvent,
) (domain.InstanceContext, error) {
	if len(currentStepDef.Outputs) == 0 {
		return instanceCtx, nil
	}
	dataToMerge := make(map[string]any)
	for _, outputParam := range currentStepDef.Outputs {
		switch outputParam.SourceNamespace {
		case domain.StepOutputSourceResult:
			value, ok := event.Result[outputParam.SourceParam]
			if ok {
				dataToMerge[outputParam.DestinationParam] = value
			}
		case domain.StepOutputSourceError:
			value, ok := event.Error.Details[outputParam.SourceParam]
			if ok {
				dataToMerge[outputParam.DestinationParam] = value
			}
		}
	}
	return instanceCtx.AppendMap(dataToMerge)
}

func buildStepTimeoutEvent(instanceId uuid.UUID, stepId string) *broker.SagaStepResultEvent {
	return &broker.SagaStepResultEvent{
		Ref: broker.SagaStepRef{
			SagaId:      instanceId,
			StepName:    stepId,
			ServiceName: "ENGINE",
		},
		Status:     broker.SagaStepStatusTimeout,
		ResolvedAt: utils.Ptr(time.Now()),
	}
}
