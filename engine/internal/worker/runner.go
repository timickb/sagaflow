package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/engine/pkg/utils"
)

type Runner struct {
	cfg          domain.RunnerConfig
	instanceRepo domain.InstanceRepository
	stepRepo     domain.StepRepository
	transactor   domain.Transactor
	sagaCache    domain.SagaDefinitionCache
	handlerCache domain.HandlerCache
}

func NewRunner(
	cfg domain.RunnerConfig,
	instanceRepo domain.InstanceRepository,
	stepRepo domain.StepRepository,
	transactor domain.Transactor,
	sagaCache domain.SagaDefinitionCache,
	handlerCache domain.HandlerCache,
) *Runner {
	return &Runner{
		cfg:          cfg,
		instanceRepo: instanceRepo,
		stepRepo:     stepRepo,
		transactor:   transactor,
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
				r.runInstance(ctx, instance)
			}
		}
	}
}

func (r *Runner) runInstance(ctx context.Context, instance *domain.InstanceView) {
	log.Info().Msgf("Start instance %v step handling", instance.SagaId)

	defer func() {
		if err := r.instanceRepo.RemoveLock(ctx, instance.SagaId); err != nil {
			log.Error().Err(err).Msgf("Failed to remove lock from instance %v", instance.SagaId)
		}
	}()

	sagaDef, found := r.sagaCache.GetSagaDefinition(domain.SagaDefinitionHeader{
		Name:    instance.SagaName,
		Version: instance.SagaVersion,
	})
	if !found {
		log.Error().Msgf("Saga %s:%d does not exist", instance.SagaName, instance.SagaVersion)
		r.failInstance(ctx, instance.SagaId, domain.InstanceFailReasonSagaNotFound, nil)
		return
	}

	switch instance.Status {
	case domain.InstanceStatusPending:
		// 1. Должен существовать шаг, указанный в saga.start
		startStepDef := utils.Find(sagaDef.Steps, func(step *domain.DefinitionStep) bool {
			return step.Id == sagaDef.StartStepId
		})
		if startStepDef == nil {
			log.Error().Msgf("First step %s does not declared in DSL, abort saga", sagaDef.StartStepId)
			r.failInstance(ctx, instance.SagaId, domain.InstanceFailReasonStepNotFound, nil)
			return
		}
		// 2. Выполнение первого шага
		r.executeStep(ctx, instance, startStepDef)
	case domain.InstanceStatusRunning:
		// 1. Должен существовать шаг, сохраненный в current_step_name
		if utils.IsStrNilOrEmpty(instance.CurrentStepName) {
			log.Error().Msgf("Instance %s has unexpected empty current_step_name", instance.SagaId)
			r.failInstance(ctx, instance.SagaId, domain.InstanceFailReasonStepNotFound, nil)
			return
		}
		pendingStepDef := utils.Find(sagaDef.Steps, func(step *domain.DefinitionStep) bool {
			return step.Id == *instance.CurrentStepName
		})
		if pendingStepDef == nil {
			log.Error().Msgf(
				"Instance %s attempts to call undeclared step %s",
				instance.SagaId, *instance.CurrentStepName,
			)
			r.failInstance(ctx, instance.SagaId, domain.InstanceFailReasonStepNotFound, nil)
			return
		}
		// 2. Выполнение очередного шага
		r.executeStep(ctx, instance, pendingStepDef)
	case domain.InstanceStatusVerifying:
		// todo
	default:
		log.Error().Msgf("Unexpected instance %v status = %v", instance.SagaId, instance.Status)
		return
	}

	log.Info().Msgf("Finish instance %v step handling", instance.SagaId)
}

func (r *Runner) failInstance(
	ctx context.Context,
	instanceId uuid.UUID,
	failReason domain.InstanceFailReason,
	msg *string,
) {
	err := r.instanceRepo.SetFailed(ctx, instanceId, &domain.InstanceFailDto{
		ErrCode:    string(failReason),
		ErrMessage: msg,
	})
	if err != nil {
		log.Error().Err(err).Msgf("Failed to set instance %v failed (errcode=%v)", instanceId, failReason)
	}
}
