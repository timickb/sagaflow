package outbox

import (
	"context"
	"fmt"

	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/db"
)

type Repository interface {
	PushSagaStepResultEvent(ctx context.Context, event *broker.SagaStepResultEvent) error
}

type RepositoryImpl struct {
	db *db.Database
}

func NewRepository(db *db.Database) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

// PushSagaStepResultEvent - сохранить событие завершения шага саги
func (r *RepositoryImpl) PushSagaStepResultEvent(ctx context.Context, event *broker.SagaStepResultEvent) error {
	dbEvent, err := NewEventDtoFromStepResultEvent(event)
	if err != nil {
		return fmt.Errorf("create outbox event record: %w", err)
	}
	if err := r.db.WithTxSupport(ctx).Create(dbEvent).Error; err != nil {
		return fmt.Errorf("create outbox event record: %w", err)
	}
	return nil
}
