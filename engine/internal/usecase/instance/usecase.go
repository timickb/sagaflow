package instance

import "github.com/timickb/sagaflow/engine/internal/domain"

type Usecase struct {
	repo  domain.InstanceRepository
	cache domain.SagaDefinitionCache
}

func NewUsecase(repo domain.InstanceRepository, cache domain.SagaDefinitionCache) *Usecase {
	return &Usecase{
		repo:  repo,
		cache: cache,
	}
}
