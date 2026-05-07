package worker

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/utils"
)

// handleCommittedTransition Обработать переход on.committed для шагов типа action/compensate
func (r *Runner) handleCommittedTransition(
	event *broker.SagaStepResultEvent,
	sagaDef *domain.SagaDefinition,
	currentStepDef *domain.DefinitionStep,
	instance *domain.InstanceView,
	currentStep *domain.StepView,
) (result *eventHandleResult, err error) {
	now := time.Now()
	neededOutcome := domain.OutcomeCommitted
	if currentStepDef.Kind == domain.StepKindVerify {
		neededOutcome = domain.OutcomeMatched
	}
	nextStepName, ok := currentStepDef.Transitions[neededOutcome]
	if !ok {
		return NewEventHandleNoTerminalStateResult(instance.SagaId, currentStep.Name, now), nil
	}
	nextStepDef := utils.Find(sagaDef.Steps, func(s *domain.DefinitionStep) bool {
		return s.Id == nextStepName
	})
	if nextStepDef == nil {
		log.Error().Msgf("Next step %s not found for instance %v", event.Ref.StepName, instance.SagaId)
		return NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonStepNotFound), nil
	}

	transitionDto := &domain.InstanceTransitionDto{
		Id:             instance.SagaId,
		NextStepName:   nextStepName,
		ExecutionState: utils.Ptr(domain.InstanceExecutionStateRunnable),
		Status:         nextStepDef.ToInstanceStatus(instance.Status),
	}
	if nextStepDef.Delay != nil {
		transitionDto.NextExecutionAt = utils.Ptr(now.Add(*nextStepDef.Delay))
	}
	if len(event.Result) > 0 {
		var newContext domain.InstanceContext
		if instance.RuntimeContext == nil {
			newContext, err = domain.NewJsonInstanceContextFromAny(event.Result)
			if err != nil {
				return nil, fmt.Errorf("create json instance context for step %v: %w", currentStep.Name, err)
			}
		} else {
			newContext, err = instance.RuntimeContext.AppendMap(event.Result)
			if err != nil {
				return nil, fmt.Errorf("append event data to instance context, step=%v: %w", currentStep.Name, err)
			}
		}
		transitionDto.RuntimeContext = newContext
	}
	inputData, err := domain.NewStepInputContext(
		nextStepDef.Inputs,
		instance.InitialContext,
		transitionDto.RuntimeContext,
	)
	if err != nil {
		log.Error().Msgf(
			"Failed to build input data for step %s for instance %v",
			event.Ref.StepName, instance.SagaId,
		)
		return NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonBuildStepInputData), nil
	}
	return &eventHandleResult{
		InstanceTransitionDto: transitionDto,
		StepCreateDto: &domain.StepCreateDto{
			InstanceId: instance.SagaId,
			StepName:   nextStepName,
			StepOrder:  currentStep.Order + 1,
			InputData:  inputData,
		},
		StepUpdateDto: &domain.StepUpdateDto{
			InstanceId: instance.SagaId,
			StepName:   currentStepDef.Id,
			Status:     utils.Ptr(domain.StepStatusCommitted),
		},
	}, nil
}
