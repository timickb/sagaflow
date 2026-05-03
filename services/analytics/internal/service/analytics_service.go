package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/timickb/sagaflow/services/analytics/internal/clickhouse"
	"github.com/timickb/sagaflow/services/analytics/internal/domain"
)

type AnalyticsService struct {
	chRepo *clickhouse.Repository
}

func NewAnalyticsService(chRepo *clickhouse.Repository) *AnalyticsService {
	return &AnalyticsService{chRepo: chRepo}
}

// HandleOrderEvent обрабатывает событие заказа из Kafka
func (s *AnalyticsService) HandleOrderEvent(ctx context.Context, data []byte) error {
	var event domain.OrderEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal order event: %w", err)
	}

	if err := s.chRepo.InsertOrderEvent(ctx, &event); err != nil {
		return fmt.Errorf("insert order event: %w", err)
	}

	return nil
}

// RebuildOrderProjection перестраивает проекцию на основе payload
func (s *AnalyticsService) RebuildOrderProjection(ctx context.Context, payload []byte) error {
	var event domain.OrderEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	if err := s.chRepo.InsertFromPayload(ctx, event.OrderID, event.UserID, event.Status, event.Version); err != nil {
		return fmt.Errorf("insert from payload: %w", err)
	}

	return nil
}
