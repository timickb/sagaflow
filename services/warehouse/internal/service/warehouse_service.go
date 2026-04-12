package service

import (
	"github.com/timickb/sagaflow/lib/outbox"
	"github.com/timickb/sagaflow/services/warehouse/internal/domain"
	"github.com/timickb/sagaflow/services/warehouse/internal/repository"
)

const (
	serviceName = "warehouse"
)

type WarehouseService struct {
	repo       *repository.WarehouseRepository
	outboxRepo outbox.Repository
	transactor domain.Transactor
}

func NewWarehouseService(
	repo *repository.WarehouseRepository,
	outboxRepo outbox.Repository,
	transactor domain.Transactor,
) *WarehouseService {
	return &WarehouseService{
		repo:       repo,
		outboxRepo: outboxRepo,
		transactor: transactor,
	}
}
