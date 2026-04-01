package instance

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
)

func (u *Usecase) Start(ctx context.Context, dto *domain.InstanceStartDto) (uuid.UUID, error) {
	_, sagaFound := u.cache.GetSagaDefinition(domain.SagaDefinitionHeader{
		Name:    dto.SagaName,
		Version: dto.SagaVersion,
	})
	if !sagaFound {
		return uuid.Nil, errors.New("saga not found")
	}

	instanceId, err := u.repo.Create(ctx, dto)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create instance: %w", err)
	}
	// TODO: отправлять аналитику
	return instanceId, nil
}
