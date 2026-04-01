package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/engine/pkg/utils"
)

type Runner struct {
	cfg          domain.Config
	instanceRepo domain.InstanceRepository
	sagaCache    domain.SagaDefinitionCache
	handlerCache domain.HandlerCache
}

func NewRunner(
	cfg domain.Config,
	instanceRepo domain.InstanceRepository,
	sagaCache domain.SagaDefinitionCache,
	handlerCache domain.HandlerCache,
) *Runner {
	return &Runner{
		cfg:          cfg,
		instanceRepo: instanceRepo,
		sagaCache:    sagaCache,
		handlerCache: handlerCache,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	wg := sync.WaitGroup{}
	wg.Add(r.cfg.GetWorkersNum())

	for i := 0; i < r.cfg.GetWorkersNum(); i++ {
		go func(idx int) {
			defer wg.Done()
			r.startWorker(ctx, i)
		}(i)
	}

	wg.Wait()
	return nil
}

func (r *Runner) startWorker(ctx context.Context, idx int) {
	log.Info().Msgf("Starting worker %d", idx)
	workerId := fmt.Sprintf("%s-%d", r.cfg.GetHostname(), idx)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("Worker %d received context done", idx)
		default:
			instances, err := r.instanceRepo.TakeBatch(ctx, r.cfg.GetBatchSize(), r.cfg.GetLockTimeout(), workerId)
			if err != nil || len(instances) == 0 {
				if err != nil {
					log.Error().Err(err).Msgf("Worker %d failed to take batch", idx)
				}
				log.Info().Msgf("Empty instances batch, sleep (worker %d)", idx)
				time.Sleep(r.cfg.GetEmptyBatchDelay())
				continue
			}
			for _, instance := range instances {
				r.handleInstanceStep(ctx, instance)
			}
		}
	}
}

func (r *Runner) handleInstanceStep(ctx context.Context, instance *domain.InstanceView) {
	log.Info().Msgf("Start instance %v step handling", instance.SagaId)

	sagaDef, found := r.sagaCache.GetSagaDefinition(domain.SagaDefinitionHeader{
		Name:    instance.SagaName,
		Version: instance.SagaVersion,
	})
	if !found {
		log.Error().Msgf("Saga %s:%d does not exist", instance.SagaName, instance.SagaVersion)
		// TODO: пометить инстанс failed
		return
	}

	switch instance.Status {
	case domain.InstanceStatusPending:
		startStepDef := utils.Find(sagaDef.Steps, func(step *domain.DefinitionStep) bool {
			return step.Id == sagaDef.StartStepId
		})
		if startStepDef == nil {
			log.Error().Msgf("First step %s does not declared in DSL, abort saga", sagaDef.StartStepId)
			// TODO: пометить инстанс failed
			return
		}
		r.executeStep(ctx, instance, startStepDef)
	case domain.InstanceStatusRunning, domain.InstanceStatusCompensating:

	case domain.InstanceStatusVerifying:
	// выполнить сверку данных
	default:
		log.Error().Msgf("Unexpected instance %v status = %v", instance.SagaId, instance.Status)
		return
	}

	log.Info().Msgf("Finish instance %v step handling", instance.SagaId)
}

func (r *Runner) executeStep(ctx context.Context, instance *domain.InstanceView, stepDef *domain.DefinitionStep) {
	switch stepDef.Kind {
	case domain.StepKindAction:
		r.callHandler(ctx, instance, stepDef)
	}
}

func (r *Runner) callHandler(ctx context.Context, instance *domain.InstanceView, stepDef *domain.DefinitionStep) {
	if stepDef.Handler == nil {
		log.Error().Msgf(
			"Unexpected nil handler in saga %s:%d for step %s",
			instance.SagaName, instance.SagaId, stepDef.Id,
		)
		return
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
