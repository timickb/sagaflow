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
	}
)

// ApplyStepResult - обработать событие завершения обработки шага.
// Либо фиксирует новое состояние экземпляра в БД, либо возвращает ошибку
// для последующего ретрая брокером.
func (r *Runner) ApplyStepResult(ctx context.Context, event *broker.SagaStepResultEvent) (err error) {
	var (
		workerId = fmt.Sprintf("consumer-%s", r.cfg.GetHostname())
		result   *eventHandleResult
	)

	// 1. Найти экземпляр
	instance, err := r.instanceRepo.GetForEvent(ctx, event.Ref.SagaId, r.cfg.GetLockTimeout(), workerId)
	if err != nil {
		return fmt.Errorf("get instance by id: %w", err)
	}

	// 2. Запланировать фиксацию перехода на новый шаг
	defer func() {
		if err != nil {
			return
		}
		dbErr := r.transactor.Transaction(ctx, func(ctx context.Context) error {
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
		log.Error().Err(dbErr).Msgf("Failed to perform transition for instance %v", instance.SagaId)
		err = dbErr
	}()

	// 3. Найти определение саги
	sagaDef, found := r.sagaCache.GetSagaDefinition(domain.SagaDefinitionHeader{
		Name:    instance.SagaName,
		Version: instance.SagaVersion,
	})
	if !found {
		log.Error().Msgf("Saga definition %s:%d not found", instance.SagaName, instance.SagaVersion)
		result = NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonSagaNotFound)
		return nil
	}

	// 4. Найти определение текущего шага
	currentStepDef := utils.Find(sagaDef.Steps, func(s *domain.DefinitionStep) bool {
		return s.Id == event.Ref.StepName
	})
	if currentStepDef == nil {
		log.Error().Msgf("Current step %s not found for instance %v", event.Ref.StepName, instance.SagaId)
		result = NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonStepNotFound)
		return nil
	}
	if !utils.Contains(allowedStepKindsForEvent, currentStepDef.Kind) {
		log.Error().Msgf("Unexpected handled step with kind %s. Only action and compensate are possible", currentStepDef.Kind)
		result = NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonInconsistentStep)
		return nil
	}

	// 5. Вытащить состояние текущего шага
	currentStep, found, err := r.stepRepo.GetByInstanceAndName(ctx, instance.SagaId, event.Ref.StepName)
	if err != nil {
		return fmt.Errorf("get current step for instance %v: %w", instance.SagaId, err)
	}
	if !found {
		log.Error().Msgf("Current step %s not found for instance %v", event.Ref.StepName, instance.SagaId)
		result = NewEventHandleFailedResult(instance.SagaId, domain.InstanceFailReasonInconsistentStep)
		return nil
	}

	// 6. Собрать DTO перехода на следующий шаг исходя из статуса события
	switch event.Status {
	case broker.SagaStepStatusCommitted:
		result, err = r.handleCommittedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		return err
	case broker.SagaStepStatusFailed:
		result, err = r.handleFailedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		return err
	case broker.SagaStepStatusRejected:
		result, err = r.handleRejectedTransition(event, sagaDef, currentStepDef, instance, currentStep)
		return err
	default:
		result = &eventHandleResult{
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
		}
		return nil
	}
}
