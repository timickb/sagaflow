package worker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/adapter/verification_source"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/utils"
)

func (r *Runner) callVerifier(
	ctx context.Context,
	instance *domain.InstanceView,
	stepDef *domain.DefinitionStep,
) {
	if stepDef.Verifier == nil {
		log.Error().Msgf(
			"Unexpected nil verifier in saga %s:%d for step %s",
			instance.SagaName, instance.SagaId, stepDef.Id,
		)
		r.failInstance(ctx, instance.SagaId, domain.InstanceFailReasonInvalidHandler, nil)
		return
	}

	if err := r.instanceRepo.SetExecutionState(ctx, instance.SagaId, domain.InstanceExecutionStateWaitingEvent); err != nil {
		log.Error().Err(err).Msgf("Unable to set WAITING_EVENT state for instance %v", instance.SagaId)
		return
	}

	verifierParams, err := r.verifiers.GetVerificationSourceParams(stepDef.Verifier.Datasource)
	if err != nil {
		log.Error().Msgf(
			"Failed to read datasource params in saga %s:%d for step %s",
			instance.SagaName, instance.SagaId, stepDef.Id,
		)
		r.failInstance(ctx, instance.SagaId, domain.InstanceFailReasonInvalidHandler, nil)
		return
	}

	source, err := verification_source.NewVerificationSource(verifierParams)
	if err != nil {
		// Не удалось достучаться до источника -> retriable ошибка
		pubErr := r.publisher.Publish(ctx, buildVerificationCallFailedEvent(instance.SagaId, stepDef.Id, err.Error()))
		if pubErr != nil {
			log.Error().Err(pubErr).Msgf(
				"Failed to publish failed verify step result (%s) for instance %v",
				err.Error(),
				instance.SagaId,
			)
		}
		return
	}
	result, err := source.Verify(ctx, &domain.VerificationRequest{
		Query:          stepDef.Verifier.Query,
		Expected:       stepDef.Verifier.Expect,
		InitialContext: instance.InitialContext,
		RuntimeContext: instance.RuntimeContext,
		Timeout:        stepDef.Timeout,
	})
	if err != nil {
		// Не удалось достучаться до источника -> retriable ошибка
		pubErr := r.publisher.Publish(ctx, buildVerificationCallFailedEvent(instance.SagaId, stepDef.Id, err.Error()))
		if pubErr != nil {
			log.Error().Err(pubErr).Msgf(
				"Failed to publish failed verify step result (%s) for instance %v",
				err.Error(),
				instance.SagaId,
			)
		}
		return
	}

	pubErr := r.publisher.Publish(ctx, buildVerificationResultEvent(instance.SagaId, stepDef.Id, result.Verified))
	if pubErr != nil {
		log.Error().Err(pubErr).Msgf(
			"Failed to publish failed verify step result (%s) for instance %v",
			err.Error(),
			instance.SagaId,
		)
	}
}

func buildVerificationCallFailedEvent(instanceId uuid.UUID, stepId string, errText string) *broker.SagaStepResultEvent {
	return &broker.SagaStepResultEvent{
		Ref: broker.SagaStepRef{
			SagaId:      instanceId,
			StepName:    stepId,
			ServiceName: "ENGINE",
		},
		Status:     broker.SagaStepStatusFailed,
		ResolvedAt: utils.Ptr(time.Now()),
		Error: &broker.ErrorInfo{
			Code:      "VERIFICATION_SOURCE_CALL_FAILED", // todo: вынести
			Message:   errText,
			Retriable: true,
		},
	}
}

func buildVerificationResultEvent(instanceId uuid.UUID, stepId string, matched bool) *broker.SagaStepResultEvent {
	status := broker.SagaStepStatusCommitted
	if !matched {
		status = broker.SagaStepStatusFailed
	}
	return &broker.SagaStepResultEvent{
		Ref: broker.SagaStepRef{
			SagaId:      instanceId,
			StepName:    stepId,
			ServiceName: "ENGINE",
		},
		Status:     status,
		ResolvedAt: utils.Ptr(time.Now()),
	}
}
