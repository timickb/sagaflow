package worker

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/utils"
)

// handleRejectedTransition Обработать переход on.rejected для шагов типа action/compensate/reconcile
// или по on.unmatched для шагов типа verify
func (r *Runner) handleRejectedTransition(
	event *broker.SagaStepResultEvent,
	sagaDef *domain.SagaDefinition,
	currentStepDef *domain.DefinitionStep,
	instance *domain.InstanceView,
	currentStep *domain.StepView,
) (result *eventHandleResult, err error) {
	var (
		now                = time.Now()
		currentStepErrData domain.InstanceContext
		instanceErrMessage *string
	)
	if event.Error != nil {
		instanceErrMessage = utils.Ptr(event.Error.String())
		currentStepErrData, err = domain.NewJsonInstanceContextFromAny(event.Error)
		if err != nil {
			return nil, fmt.Errorf("create instance context from error data struct: %w", err)
		}
	}

	// Поиск следующего шага для исхода
	neededOutcome := domain.OutcomeRejected
	if currentStepDef.Kind == domain.StepKindVerify {
		neededOutcome = domain.OutcomeUnmatched
	}
	nextStepName, ok := currentStepDef.Transitions[neededOutcome]
	if !ok {
		// Перехода нет -> завершить инстанс со статусом INCONSISTENT
		return NewEventHandleNoTerminalStateResult(instance.SagaId, currentStep.Name), nil
	}
	nextStepDef := utils.Find(sagaDef.Steps, func(s *domain.DefinitionStep) bool {
		return s.Id == nextStepName
	})
	if nextStepDef == nil {
		log.Error().Msgf("Next step %s not found for instance %v", nextStepName, instance.SagaId)
		return NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonStepNotFound), nil
	}

	// Сбор входных данных для следующего шага
	inputData, err := domain.NewStepInputContext(
		nextStepDef.Inputs,
		instance.InitialContext,
		instance.RuntimeContext,
	)
	if err != nil {
		log.Error().Msgf(
			"Failed to build input data for step %s for instance %v",
			event.Ref.StepName, instance.SagaId,
		)
		return NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonBuildStepInputData), nil
	}

	// Применение выходных данных текущего шага к контексту сценария
	newRuntimeContext, err := mergeStepOutputToContext(instance.RuntimeContext, currentStepDef, event)
	if err != nil {
		return NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonApplyStepOutputData), nil
	}

	// если следующий шаг терминальный - сразу ставим ему COMMITTED статус
	var newStepStatus *domain.StepStatus
	if nextStepDef.Kind == domain.StepKindTerminal {
		newStepStatus = utils.Ptr(domain.StepStatusCommitted)
	}
	instanceTransitionDto := &domain.InstanceTransitionDto{
		Id:             instance.SagaId,
		NextStepName:   nextStepDef.Id,
		Status:         nextStepDef.ToInstanceStatus(instance.Status),
		ExecutionState: utils.Ptr(domain.InstanceExecutionStateRunnable),
		RuntimeContext: newRuntimeContext,
	}
	if nextStepDef.Delay != nil {
		instanceTransitionDto.NextExecutionAt = utils.Ptr(now.Add(*nextStepDef.Delay))
	}
	if instanceErrMessage != nil {
		instanceTransitionDto.ErrCode = utils.Ptr(string(domain.InstanceErrorCodeHandler))
		instanceTransitionDto.ErrMessage = instanceErrMessage
	}
	return &eventHandleResult{
		InstanceTransitionDto: instanceTransitionDto,
		StepCreateDto: &domain.StepCreateDto{
			InstanceId: instance.SagaId,
			StepName:   nextStepDef.Id,
			StepOrder:  currentStep.Order + 1,
			InputData:  inputData,
			Status:     newStepStatus,
		},
		StepUpdateDto: &domain.StepUpdateDto{
			InstanceId: instance.SagaId,
			StepName:   currentStep.Name,
			Status:     utils.Ptr(domain.StepStatusFailed),
			ErrorData:  currentStepErrData,
		},
	}, nil
}
