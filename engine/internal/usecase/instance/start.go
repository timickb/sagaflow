package instance

import (
	"context"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
)

func (u *Usecase) Start(ctx context.Context, dto *domain.InstanceStartDto) (uuid.UUID, error) {
	// TODO: implement
	panic("implement me")
}
