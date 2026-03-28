package instance

import (
	"context"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
)

func (u *Usecase) GetInfo(ctx context.Context, sagaId uuid.UUID) (*domain.InstanceView, error) {
	// TODO: implement
	panic("implement me")
}
