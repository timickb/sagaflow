package instance

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
)

func (u *Usecase) Start(ctx context.Context, dto *domain.InstanceStartDto) (instanceId uuid.UUID, err error) {
	sagaDef, sagaFound := u.cache.GetSagaDefinition(domain.SagaDefinitionHeader{
		Name:    dto.SagaName,
		Version: dto.SagaVersion,
	})
	if !sagaFound {
		return uuid.Nil, errors.New("saga not found")
	}

	startStepDef, ok := sagaDef.StepById[sagaDef.StartStepId]
	if !ok {
		return uuid.Nil, errors.New("start step not declared in saga")
	}
	dto.StartStepName = sagaDef.StartStepId

	// TODO: собрать InputData из startStepDef.Input и sagaDef.Inputs

	err = u.transactor.Transaction(ctx, func(ctx context.Context) error {
		instanceId, err = u.repo.Create(ctx, dto)
		if err != nil {
			return fmt.Errorf("failed to save instance: %w", err)
		}
		_, err = u.stepRepo.Create(ctx, &domain.StepCreateDto{
			InstanceId: instanceId,
			StepName:   sagaDef.StartStepId,
			StepOrder:  1,
			InputData:  nil,
		})
		if err != nil {
			return fmt.Errorf("failed to save start step: %w", err)
		}

		return nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to start instance: %w", err)
	}

	// TODO: отправлять аналитику
	return instanceId, nil
}
