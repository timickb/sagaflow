package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/utils"
)

type Runner struct {
	cfg          domain.RunnerConfig
	handlers     domain.HandlersConfig
	verifiers    domain.VerificationSourcesConfig
	instanceRepo domain.InstanceRepository
	stepRepo     domain.StepRepository
	transactor   domain.Transactor
	sagaCache    domain.SagaDefinitionCache
	stepHandler  domain.StepHandler
	publisher    *broker.KafkaStepResultWriter
}

func NewRunner(
	cfg domain.RunnerConfig,
	handlersCfg domain.HandlersConfig,
	verifiersCfg domain.VerificationSourcesConfig,
	instanceRepo domain.InstanceRepository,
	stepRepo domain.StepRepository,
	transactor domain.Transactor,
	sagaCache domain.SagaDefinitionCache,
	stepHandler domain.StepHandler,
	publisher *broker.KafkaStepResultWriter,
) *Runner {
	return &Runner{
		cfg:          cfg,
		handlers:     handlersCfg,
		verifiers:    verifiersCfg,
		instanceRepo: instanceRepo,
		stepRepo:     stepRepo,
		transactor:   transactor,
		sagaCache:    sagaCache,
		stepHandler:  stepHandler,
		publisher:    publisher,
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

	// 1. Должен был задекларирован шаг, сохраненный в current_step_name
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
	// 2. Он должен быть сохранен в saga_steps
	pendingStep, exists, err := r.stepRepo.GetByInstanceAndName(ctx, instance.SagaId, *instance.CurrentStepName)
	if err != nil {
		log.Error().Err(err).Msgf(
			"Failed to fetch pending step from db: instance=%v, step=%s",
			instance.SagaId, *instance.CurrentStepName,
		)
		return
	}
	if !exists {
		log.Error().Msgf(
			"Instance %s attempts to call declared step %s, but it wasn't created",
			instance.SagaId, *instance.CurrentStepName,
		)
		r.failInstance(ctx, instance.SagaId, domain.InstanceFailReasonStepNotFound, nil)
		return
	}
	// 3. Выполнение очередного шага
	r.executeStep(ctx, instance, pendingStepDef, pendingStep)

	log.Info().Msgf("Finish instance %v step handling", instance.SagaId)
}

func (r *Runner) executeStep(
	ctx context.Context,
	instance *domain.InstanceView,
	stepDef *domain.DefinitionStep,
	step *domain.StepView,
) {
	switch stepDef.Kind {
	case domain.StepKindAction, domain.StepKindCompensate, domain.StepKindReconcile:
		r.callHandler(ctx, instance, stepDef, step)
	case domain.StepKindVerify:
		r.callVerifier(ctx, instance, stepDef)
	}
}

func (r *Runner) failInstance(
	ctx context.Context,
	instanceId uuid.UUID,
	failReason domain.InstanceFailReason,
	msg *string,
) {
	err := r.instanceRepo.Terminate(ctx, instanceId, &domain.InstanceTerminateDto{
		ErrCode:    utils.Ptr(string(failReason)),
		ErrMessage: msg,
		Status:     domain.InstanceStatusFailed,
	})
	if err != nil {
		log.Error().Err(err).Msgf("Failed to set instance %v failed (errcode=%v)", instanceId, failReason)
	}
}

func (r *Runner) finishInstance(
	ctx context.Context,
	instanceId uuid.UUID,
	status domain.InstanceStatus,
) {
	err := r.instanceRepo.Terminate(ctx, instanceId, &domain.InstanceTerminateDto{
		Status: status,
	})
	if err != nil {
		log.Error().Err(err).Msgf("Failed to set instance %v finished", instanceId)
	}
}
