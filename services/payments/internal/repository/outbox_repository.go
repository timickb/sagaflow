package repository

import (
	"github.com/timickb/sagaflow/lib/outbox"
)

type OutboxRepository struct {
	*outbox.RepositoryImpl
}

func NewOutboxRepository(impl *outbox.RepositoryImpl) *OutboxRepository {
	return &OutboxRepository{RepositoryImpl: impl}
}
