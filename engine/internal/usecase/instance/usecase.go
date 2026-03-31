package instance

import "github.com/timickb/sagaflow/engine/internal/domain"

type Usecase struct {
	repo  domain.InstanceRepository
	cache domain.SagaCache
}

func NewUsecase(repo domain.InstanceRepository, cache domain.SagaCache) *Usecase {
	return &Usecase{
		repo:  repo,
		cache: cache,
	}
}
