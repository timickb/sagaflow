package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/domain"
)

type Runner struct {
	cfg          domain.Config
	instanceRepo domain.InstanceRepository
	sagaCache    domain.SagaCache
}

func NewRunner(
	cfg domain.Config,
	instanceRepo domain.InstanceRepository,
	sagaCache domain.SagaCache,
) *Runner {
	return &Runner{
		cfg:          cfg,
		instanceRepo: instanceRepo,
		sagaCache:    sagaCache,
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
				r.handleInstanceStep(ctx, instance)
			}
		}
	}
}

func (r *Runner) handleInstanceStep(ctx context.Context, instance *domain.InstanceView) {
	log.Info().Msgf("Start instance %v step handling", instance.SagaId)

	sagaView, found := r.sagaCache.GetSagaView(domain.SagaHeader{
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

	case domain.InstanceStatusRunning, domain.InstanceStatusCompensating:
	// отправить grpc запрос в обработчик
	case domain.InstanceStatusVerifying:
	// выполнить сверку данных
	default:
		log.Error().Msgf("Unexpected instance %v status = %v", instance.SagaId, instance.Status)
		return
	}

	log.Info().Msgf("Finish instance %v step handling", instance.SagaId)
}
