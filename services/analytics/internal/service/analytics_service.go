package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/services/analytics/internal/clickhouse"
	"github.com/timickb/sagaflow/services/analytics/internal/client"
	"github.com/timickb/sagaflow/services/analytics/internal/domain"
)

type rebuildOrderPayload struct {
	OrderId uuid.UUID `json:"order_id"`
}

type AnalyticsService struct {
	chRepo           *clickhouse.Repository
	paymentsClient   *client.PaymentsClient
	stepResultWriter broker.StepResultWriter
}

func NewAnalyticsService(chRepo *clickhouse.Repository, paymentsClient *client.PaymentsClient, stepResultWriter broker.StepResultWriter) *AnalyticsService {
	return &AnalyticsService{
		chRepo:           chRepo,
		paymentsClient:   paymentsClient,
		stepResultWriter: stepResultWriter,
	}
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

// RebuildOrderProjectionMeta - метаданные для публикации результата шага
type RebuildOrderProjectionMeta struct {
	SagaId   uuid.UUID
	StepName string
}

// RebuildOrderProjection перестраивает проекцию на основе текущего состояния заказа
func (s *AnalyticsService) RebuildOrderProjection(ctx context.Context, payload []byte, meta RebuildOrderProjectionMeta) error {
	var data rebuildOrderPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("unmarshal rebuild order payload: %w", err)
	}
	orderId := data.OrderId

	// Вызов ручки PaymentsService для получения сущности заказа
	resp, err := s.paymentsClient.GetOrderById(ctx, orderId.String())
	if err != nil {
		s.publishStepResult(ctx, meta, broker.SagaStepStatusFailed, nil, err)
		return fmt.Errorf("get order by id from payments service: %w", err)
	}

	// Маппинг ответа в domain.OrderEvent
	userId, err := uuid.Parse(resp.GetUserId())
	if err != nil {
		s.publishStepResult(ctx, meta, broker.SagaStepStatusFailed, nil, err)
		return fmt.Errorf("parse user_id: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339, resp.GetCreatedAt())
	if err != nil {
		s.publishStepResult(ctx, meta, broker.SagaStepStatusFailed, nil, err)
		return fmt.Errorf("parse created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, resp.GetUpdatedAt())
	if err != nil {
		s.publishStepResult(ctx, meta, broker.SagaStepStatusFailed, nil, err)
		return fmt.Errorf("parse updated_at: %w", err)
	}

	event := domain.OrderEvent{
		OrderId:     orderId,
		UserId:      userId,
		Status:      resp.GetStatus(),
		Currency:    resp.GetCurrency(),
		TotalAmount: float64(resp.GetTotalAmount()),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Version:     uint32(resp.GetVersion()),
	}

	if err = s.chRepo.InsertOrderEvent(ctx, &event); err != nil {
		s.publishStepResult(ctx, meta, broker.SagaStepStatusFailed, nil, err)
		return fmt.Errorf("insert from payload: %w", err)
	}

	s.publishStepResult(ctx, meta, broker.SagaStepStatusCommitted, nil, nil)
	return nil
}

// publishStepResult публикует результат выполнения шага в топик saga.step.result
func (s *AnalyticsService) publishStepResult(ctx context.Context, meta RebuildOrderProjectionMeta, status broker.SagaStepResultStatus, result map[string]any, stepErr error) {
	if s.stepResultWriter == nil {
		return
	}

	now := time.Now()
	event := &broker.SagaStepResultEvent{
		Ref: broker.SagaStepRef{
			SagaId:      meta.SagaId,
			StepName:    meta.StepName,
			ServiceName: "analytics",
		},
		Worker: broker.WorkerInfo{
			InstanceId: getInstanceId(),
			Hostname:   getHostname(),
		},
		Status:     status,
		ResolvedAt: &now,
		Result:     result,
	}

	if stepErr != nil {
		event.Error = &broker.ErrorInfo{
			Code:      "ANALYTICS_REBUILD_ERROR",
			Message:   stepErr.Error(),
			Retriable: false,
		}
	}

	if err := s.stepResultWriter.Publish(ctx, event); err != nil {
		fmt.Printf("failed to publish step result: %v\n", err)
	}
}

func getInstanceId() string {
	hostname, _ := os.Hostname()
	return hostname
}

func getHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}
