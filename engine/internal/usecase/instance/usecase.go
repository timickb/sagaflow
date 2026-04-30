package instance

import "github.com/timickb/sagaflow/engine/internal/domain"

type Usecase struct {
	repo       domain.InstanceRepository
	stepRepo   domain.StepRepository
	transactor domain.Transactor
	cache      domain.SagaDefinitionCache
}

func NewUsecase(
	repo domain.InstanceRepository,
	stepRepo domain.StepRepository,
	transactor domain.Transactor,
	cache domain.SagaDefinitionCache,
) *Usecase {
	return &Usecase{
		repo:       repo,
		stepRepo:   stepRepo,
		transactor: transactor,
		cache:      cache,
	}
}
