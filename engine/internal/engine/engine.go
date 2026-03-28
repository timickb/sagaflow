package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/domain"
)

type Runner struct {
	cfg          domain.Config
	instanceRepo domain.InstanceRepository
}

func NewRunner(cfg domain.Config, instanceRepo domain.InstanceRepository) *Runner {
	return &Runner{
		cfg:          cfg,
		instanceRepo: instanceRepo,
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
			if err != nil {
				log.Error().Err(err).Msgf("Worker %d failed to take batch", idx)
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
	// TODO: implement
	log.Info().Msgf("Finish instance %v step handling", instance.SagaId)
}
