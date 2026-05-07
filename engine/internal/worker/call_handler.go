package worker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/utils"
)

func (r *Runner) callHandler(
	ctx context.Context,
	instance *domain.InstanceView,
	stepDef *domain.DefinitionStep,
	step *domain.StepView,
) {
	// 1. Найти дефиницию обработчика
	if stepDef.Handler == nil {
		log.Error().Msgf(
			"Unexpected nil handler in saga %s:%d for step %s",
			instance.SagaName, instance.SagaId, stepDef.Id,
		)
		r.failInstance(ctx, instance.SagaId, domain.InstanceFailReasonInvalidHandler, nil)
		return
	}

	// 2. Установить WAITING_EVENT экземпляру
	if err := r.instanceRepo.SetExecutionState(ctx, instance.SagaId, domain.InstanceExecutionStateWaitingEvent); err != nil {
		log.Error().Err(err).Msgf("Unable to set WAITING_EVENT state for instance %v", instance.SagaId)
		return
	}

	// 3. Вызвать обработчик
	callResult, err := r.stepHandler.Call(ctx, &domain.CallHandlerRequest{
		Service:        stepDef.Handler.Service,
		Method:         stepDef.Handler.Method,
		SagaInstanceId: instance.SagaId,
		StepId:         stepDef.Id,
		Attempt:        step.Attempt + 1,
		InputData:      step.InputData,
	})
	if err != nil {
		log.Error().Err(err).Msgf("Failed to call handler for step %s in saga %s", stepDef.Id, instance.SagaName)
		pubErr := r.publisher.Publish(ctx, buildCallFailedEvent(instance.SagaId, stepDef.Id, err.Error()))
		if pubErr != nil {
			log.Error().Err(pubErr).Msgf(
				"Failed to publish failed step result (%s) for instance %v",
				err.Error(),
				instance.SagaId,
			)
		}
		return
	}
	switch callResult.Status {
	case domain.CallHandlerResultUnprocessable:
		errText := "error"
		if callResult.ErrorData != nil {
			errText = *callResult.ErrorData
		}
		log.Error().Msgf(
			"Failed to handle step %s in saga %s: handler responsed %s, abort",
			stepDef.Id, instance.SagaName, errText,
		)
		r.failInstance(ctx, instance.SagaId, domain.InstanceFailReasonInvalidHandler, nil)
	case domain.CallHandlerResultHandlerNotFound:
		log.Error().Msgf("No handler for step %s in saga %s, abort", stepDef.Id, instance.SagaName)
		r.failInstance(ctx, instance.SagaId, domain.InstanceFailReasonHandlerNotFound, nil)
	case domain.CallHandlerResultSuccess:
		log.Info().Msgf("Successfully invoked handler for step %s in saga %s", stepDef.Id, instance.SagaName)
	}
}

func buildCallFailedEvent(instanceId uuid.UUID, stepId string, errText string) *broker.SagaStepResultEvent {
	return &broker.SagaStepResultEvent{
		Ref: broker.SagaStepRef{
			SagaId:      instanceId,
			StepName:    stepId,
			ServiceName: "ENGINE",
		},
		Status:     broker.SagaStepStatusFailed,
		ResolvedAt: utils.Ptr(time.Now()),
		Error: &broker.ErrorInfo{
			Code:      "RPC_CALL_FAILED", // todo: вынести
			Message:   errText,
			Retriable: true,
		},
	}
}
