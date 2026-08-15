package worker

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/utils"
)

// stepDisposition описывает, как именно обработчик исхода распоряжается
// текущим шагом и инстансом — то, чем отличаются друг от друга переходы
// committed/failed/rejected/timeout.
type stepDisposition struct {
	currentStepStatus        domain.StepStatus      // статус, в который перевести текущий шаг
	errorData                domain.InstanceContext // данные ошибки текущего шага (nil, если ошибки нет)
	incrementReconcileCycles bool                   // увеличить счетчик циклов reconcile
	instanceErrCode          *string                // код ошибки на переходе инстанса
	instanceErrMsg           *string                // сообщение об ошибке на переходе инстанса
}

// resolveOutcome возвращает исход перехода с учетом того, что для шагов
// типа verify используются специальные исходы (matched/unmatched).
func resolveOutcome(kind domain.StepKind, base, verifyVariant domain.StepOutcome) domain.StepOutcome {
	if kind == domain.StepKindVerify {
		return verifyVariant
	}
	return base
}

// advanceOnOutcome — общий путь перехода экземпляра саги на следующий шаг
// по исходу outcome. Различия между конкретными исходами передаются через disp.
func (r *Runner) advanceOnOutcome(
	event *broker.SagaStepResultEvent,
	sagaDef *domain.SagaDefinition,
	currentStepDef *domain.DefinitionStep,
	instance *domain.InstanceView,
	currentStep *domain.StepView,
	outcome domain.StepOutcome,
	disp stepDisposition,
) (*eventHandleResult, error) {
	now := time.Now()

	nextStepName, ok := currentStepDef.Transitions[outcome]
	if !ok {
		// Перехода для исхода нет -> завершить инстанс нетерминальным состоянием
		return NewEventHandleNoTerminalStateResult(instance.SagaId, currentStep.Name), nil
	}
	nextStepDef := utils.Find(sagaDef.Steps, func(s *domain.DefinitionStep) bool {
		return s.Id == nextStepName
	})
	if nextStepDef == nil {
		log.Error().Msgf("Next step %s not found for instance %v", nextStepName, instance.SagaId)
		return NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonStepNotFound), nil
	}

	// Применение выходных данных текущего шага к контексту сценария
	newRuntimeContext, err := mergeStepOutputToContext(instance.RuntimeContext, currentStepDef, event)
	if err != nil {
		return NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonApplyStepOutputData), nil
	}

	// Сбор входных данных для следующего шага из обновленного контекста
	inputData, err := domain.NewStepInputContext(
		nextStepDef.Inputs,
		instance.InitialContext,
		newRuntimeContext,
	)
	if err != nil {
		log.Error().Msgf(
			"Failed to build input data for step %s for instance %v",
			nextStepDef.Id, instance.SagaId,
		)
		return NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonBuildStepInputData), nil
	}

	transition := &domain.InstanceTransitionDto{
		Id:             instance.SagaId,
		NextStepName:   nextStepDef.Id,
		Status:         nextStepDef.ToInstanceStatus(instance.Status),
		ExecutionState: utils.Ptr(domain.InstanceExecutionStateRunnable),
		RuntimeContext: newRuntimeContext,
		ErrCode:        disp.instanceErrCode,
		ErrMessage:     disp.instanceErrMsg,
	}
	if nextStepDef.Delay != nil {
		transition.NextExecutionAt = utils.Ptr(now.Add(*nextStepDef.Delay))
	}

	// если следующий шаг терминальный - сразу ставим ему COMMITTED статус
	var nextStepStatus *domain.StepStatus
	if nextStepDef.Kind == domain.StepKindTerminal {
		nextStepStatus = utils.Ptr(domain.StepStatusCommitted)
	}

	return &eventHandleResult{
		InstanceTransitionDto: transition,
		StepCreateDto: &domain.StepCreateDto{
			InstanceId: instance.SagaId,
			StepName:   nextStepDef.Id,
			StepOrder:  currentStep.Order + 1,
			InputData:  inputData,
			Status:     nextStepStatus,
		},
		StepUpdateDto: &domain.StepUpdateDto{
			InstanceId:               instance.SagaId,
			StepName:                 currentStep.Name,
			Status:                   utils.Ptr(disp.currentStepStatus),
			ErrorData:                disp.errorData,
			IncrementReconcileCycles: disp.incrementReconcileCycles,
		},
	}, nil
}
