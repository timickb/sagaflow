package worker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/utils"
	"github.com/timickb/sagaflow/proto/gen/go/sagaflow"
)

func (r *Runner) callHandler(
	ctx context.Context,
	instance *domain.InstanceView,
	stepDef *domain.DefinitionStep,
	step *domain.StepView,
) {
	if stepDef.Handler == nil {
		log.Error().Msgf(
			"Unexpected nil handler in saga %s:%d for step %s",
			instance.SagaName, instance.SagaId, stepDef.Id,
		)
		r.failInstance(ctx, instance.SagaId, domain.InstanceFailReasonInvalidHandler, nil)
		return
	}

	if err := r.instanceRepo.SetExecutionState(ctx, instance.SagaId, domain.InstanceExecutionStateWaitingEvent); err != nil {
		log.Error().Err(err).Msgf("Unable to set WAITING_EVENT state for instance %v", instance.SagaId)
		return
	}

	req := &sagaflow.HandleRequest{
		Meta: &sagaflow.StepExecutionMeta{
			SagaId:         instance.SagaId.String(),
			StepId:         stepDef.Id,
			Action:         stepDef.Handler.Method,
			Attempt:        int32(step.Attempt + 1),
			IdempotencyKey: "", // todo
		},
		Payload: step.InputData.GetRaw(),
	}
	conn, found := r.handlers.GetHandlerConnection(stepDef.Handler.Service)
	if !found || conn == nil {
		log.Error().Msgf("No handler for step %s in saga %s, abort", stepDef.Id, instance.SagaName)
		r.failInstance(ctx, instance.SagaId, domain.InstanceFailReasonHandlerNotFound, nil)
		return
	}
	resp, err := sagaflow.NewStepHandlerServiceClient(conn).Handle(ctx, req)
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
	if !resp.Success {
		// синхронный ответ с ошибкой не может быть retriable
		// TODO: не всегда!
		errText := "error"
		if resp.Error != nil {
			errText = *resp.Error
		}
		log.Error().Msgf(
			"Failed to handle step %s in saga %s: handler responsed %s, abort",
			stepDef.Id, instance.SagaName, errText,
		)
		r.failInstance(ctx, instance.SagaId, domain.InstanceFailReasonInvalidHandler, nil)
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
