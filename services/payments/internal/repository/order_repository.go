package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/lib/db"
	"github.com/timickb/sagaflow/services/payments/internal/domain"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *db.Database
}

func NewOrderRepository(db *db.Database) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	var order domain.Order
	if err := r.db.WithContext(ctx).First(&order, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *OrderRepository) UpdateDetails(ctx context.Context, id uuid.UUID, details []byte) error {
	return r.db.WithContext(ctx).
		Model(&domain.Order{}).
		Where("id = ?", id).
		Update("details", details).Error
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	return r.db.WithContext(ctx).
		Model(&domain.Order{}).
		Where("id = ?", id).
		Update("status", status).Error
}
