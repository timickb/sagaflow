package worker

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/proto/gen/go/sagaflow"
)

func (r *Runner) callHandler(ctx context.Context, instance *domain.InstanceView, stepDef *domain.DefinitionStep) {
	if stepDef.Handler == nil {
		log.Error().Msgf(
			"Unexpected nil handler in saga %s:%d for step %s",
			instance.SagaName, instance.SagaId, stepDef.Id,
		)
		return
	}
	_ = &sagaflow.StepExecutionMeta{
		SagaId:         instance.SagaId.String(),
		StepId:         stepDef.Id,
		Worker:         stepDef.Handler.Method,
		Attempt:        0,  // todo
		IdempotencyKey: "", // todo
	}
	// lookup handler
	//conn, found := r.handlerCache.GetHandlerClient(stepDef.Handler)
	//if !found || conn == nil {
	//	log.Error().Msgf("No handler for step %s in saga %s, abort", stepDef.Id, instance.SagaName)
	//	// TODO: пометить инстанс failed
	//	return
	//}
	//fullMethod := fmt.Sprintf("/%s/%s", stepDef.Handler.Service, stepDef.Handler.Method)
	//
	//callCtx, cancel := context.WithTimeout(ctx, stepDef.Timeout)
	//defer cancel()
	//
	//if err := conn.Invoke(callCtx, fullMethod, req, resp); err != nil {
	//	return fmt.Errorf("invoke grpc method %s: %w", fullMethod, err)
	//}

	return
}
