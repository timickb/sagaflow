package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/lib/db"
	"github.com/timickb/sagaflow/services/payments/internal/domain"
)

type PaymentRepository struct {
	db *db.Database
}

func NewPaymentRepository(db *db.Database) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == domain.PaymentStatusCaptured {
		now := time.Now()
		updates["captured_at"] = now
	}

	return r.db.WithContext(ctx).
		Model(&domain.Payment{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.Payment, error) {
	var payments []domain.Payment
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}
