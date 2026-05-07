package worker

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/utils"
)

var (
	allowedStepKindsForEvent = []domain.StepKind{
		domain.StepKindAction,
		domain.StepKindCompensate,
		domain.StepKindReconcile,
		domain.StepKindVerify,
	}
)

// ApplyStepResult - обработать событие завершения обработки шага.
// Либо фиксирует новое состояние экземпляра в БД, либо возвращает ошибку
// для последующего ретрая брокером.
func (r *Runner) ApplyStepResult(ctx context.Context, event *broker.SagaStepResultEvent) error {
	err := r.transactor.Transaction(ctx, func(ctx context.Context) error {
		instance, iErr := r.instanceRepo.GetForEvent(ctx, event.Ref.SagaId)
		if iErr != nil {
			return fmt.Errorf("get instance by id: %w", iErr)
		}
		if instance == nil {
			log.Info().Msgf("instance %s is not available for event applying, skip it", event.Ref.SagaId.String())
			return nil
		}

		result, rErr := r.buildStepResult(ctx, event, instance)
		if rErr != nil {
			return fmt.Errorf("build step result: %w", rErr)
		}

		if result.InstanceTransitionDto != nil {
			if tErr := r.instanceRepo.MakeTransition(ctx, result.InstanceTransitionDto); tErr != nil {
				return fmt.Errorf("make transition: %w", tErr)
			}
		}
		if result.StepUpdateDto != nil {
			if sErr := r.stepRepo.Update(ctx, result.StepUpdateDto); sErr != nil {
				return fmt.Errorf("update current step: %w", sErr)
			}
		}
		if result.StepCreateDto != nil {
			if _, sErr := r.stepRepo.Create(ctx, result.StepCreateDto); sErr != nil {
				return fmt.Errorf("save new step: %w", sErr)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("perform transaction: %w", err)
	}
	return nil
}

func (r *Runner) buildStepResult(
	ctx context.Context, event *broker.SagaStepResultEvent, instance *domain.InstanceView,
) (*eventHandleResult, error) {
	// 1. Найти определение саги
	sagaDef, found := r.sagaCache.GetSagaDefinition(domain.SagaDefinitionHeader{
		Name:    instance.SagaName,
		Version: instance.SagaVersion,
	})
	if !found {
		log.Error().Msgf("Saga definition %s:%d not found", instance.SagaName, instance.SagaVersion)
		return NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonSagaNotFound), nil
	}

	// 2. Найти определение текущего шага
	currentStepDef := utils.Find(sagaDef.Steps, func(s *domain.DefinitionStep) bool {
		return s.Id == event.Ref.StepName
	})
	if currentStepDef == nil {
		log.Error().Msgf("Current step %s not found for instance %v", event.Ref.StepName, instance.SagaId)
		return NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonStepNotFound), nil
	}
	if !utils.Contains(allowedStepKindsForEvent, currentStepDef.Kind) {
		log.Error().Msgf("Unexpected handled step with kind %s. Only action and compensate are possible", currentStepDef.Kind)
		return NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonInconsistentStep), nil
	}

	// 3. Вытащить состояние текущего шага
	currentStep, found, err := r.stepRepo.GetByInstanceAndName(ctx, instance.SagaId, event.Ref.StepName)
	if err != nil {
		return nil, fmt.Errorf("get current step for instance %v: %w", instance.SagaId, err)
	}
	if !found {
		log.Error().Msgf("Current step %s not found for instance %v", event.Ref.StepName, instance.SagaId)
		return NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonInconsistentStep), nil
	}

	// 4. Собрать DTO перехода на следующий шаг исходя из статуса события
	switch event.Status {
	case broker.SagaStepStatusCommitted:
		return r.handleCommittedTransition(event, sagaDef, currentStepDef, instance, currentStep)
	case broker.SagaStepStatusFailed:
		return r.handleFailedTransition(event, sagaDef, currentStepDef, instance, currentStep)
	case broker.SagaStepStatusRejected:
		return r.handleRejectedTransition(event, sagaDef, currentStepDef, instance, currentStep)
	default:
		return &eventHandleResult{
			InstanceTransitionDto: &domain.InstanceTransitionDto{
				Id:      instance.SagaId,
				Status:  utils.Ptr(domain.InstanceStatusFailed),
				ErrCode: utils.Ptr(string(domain.InstanceFailReasonUnknownEventStatus)),
			},
			StepUpdateDto: &domain.StepUpdateDto{
				InstanceId: instance.SagaId,
				StepName:   currentStep.Name,
				Status:     utils.Ptr(domain.StepStatusFailed),
			},
		}, nil
	}
}
