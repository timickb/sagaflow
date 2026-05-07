package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/lib/db"
	"github.com/timickb/sagaflow/services/payments/internal/domain"
)

const (
	orderAnalyticsTopicName = "analytics.orders.events"
)

type AnalyticsOutboxRepository struct {
	db *db.Database
}

func NewAnalyticsOutboxRepository(db *db.Database) *AnalyticsOutboxRepository {
	return &AnalyticsOutboxRepository{db: db}
}

func (r *AnalyticsOutboxRepository) PushOrderAnalyticsEvent(ctx context.Context, dto *domain.Order) error {
	payload := &domain.OrderAnalyticsEvent{
		OrderId:     dto.ID,
		UserId:      dto.UserID,
		Status:      string(dto.Status),
		Currency:    dto.Currency,
		TotalAmount: dto.TotalAmount,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
		Version:     dto.Version,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	dbEvent := map[string]interface{}{
		"id":             uuid.New(),
		"aggregate_type": "Order",
		"aggregate_id":   dto.ID.String(),
		"event_type":     "RefreshOrderMaterialization",
		"payload":        payloadBytes,
		"created_at":     time.Now(),
	}
	err = r.db.WithTxSupport(ctx).
		Table("domain_outbox_events").
		Create(dbEvent).Error
	if err != nil {
		return fmt.Errorf("create outbox analytics event record: %w", err)
	}
	return nil
}
