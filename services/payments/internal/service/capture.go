package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/services/payments/internal/domain"
	"github.com/timickb/sagaflow/services/payments/internal/payment"
)

func (s *PaymentsService) Capture(
	ctx context.Context,
	orderID uuid.UUID,
	sagaInstanceID uuid.UUID,
	stepID string,
	items []interface{},
	userID uuid.UUID,
	amount int,
) (*CaptureResult, error) {
	var (
		result     *CaptureResult
		captureErr error
	)

	err := s.transactor.Transaction(ctx, func(ctx context.Context) error {
		// Обновляем детали заказа
		details, err := json.Marshal(items)
		if err != nil {
			captureErr = fmt.Errorf("marshal details: %w", err)
			return captureErr
		}

		if err := s.orderRepo.UpdateDetails(ctx, orderID, details); err != nil {
			captureErr = fmt.Errorf("update order details: %w", err)
			return captureErr
		}

		// Создаем запись о платеже
		paymentID := uuid.New()
		paymentRecord := &domain.Payment{
			ID:              paymentID,
			OrderID:         orderID,
			Amount:          float64(amount),
			Currency:        "RUB",
			PaymentMethod:   "card",
			PaymentProvider: "stub",
			Status:          domain.PaymentStatusPending,
		}

		if err := s.paymentRepo.Create(ctx, paymentRecord); err != nil {
			captureErr = fmt.Errorf("create payment: %w", err)
			return err
		}

		// Вызываем платежную систему
		captureResp, err := s.paymentProv.Capture(ctx, &payment.CaptureRequest{
			PaymentID:     paymentID,
			Amount:        float64(amount),
			Currency:      "RUB",
			PaymentMethod: "card",
			OrderID:       orderID,
		})
		if err != nil {
			captureErr = fmt.Errorf("payment provider error: %w", err)
			return captureErr
		}

		// Обрабатываем результат платежа
		var stepStatus broker.SagaStepResultStatus
		var errorInfo *broker.ErrorInfo

		if captureResp.Success {
			// providerPaymentID := captureResp.ProviderPaymentID
			if err := s.paymentRepo.UpdateStatus(ctx, paymentID, domain.PaymentStatusCaptured); err != nil {
				captureErr = fmt.Errorf("update payment status: %w", err)
				return captureErr
			}

			if err := s.orderRepo.UpdateStatus(ctx, orderID, domain.OrderStatusPaid); err != nil {
				captureErr = fmt.Errorf("update order status: %w", err)
				return captureErr
			}

			stepStatus = broker.SagaStepStatusCommitted
		} else {
			if err := s.paymentRepo.UpdateStatus(ctx, paymentID, domain.PaymentStatusFailed); err != nil {
				captureErr = fmt.Errorf("update payment status: %w", err)
				return captureErr
			}

			stepStatus = broker.SagaStepStatusFailed
			errorInfo = &broker.ErrorInfo{
				Code:    captureResp.ErrorCode,
				Message: captureResp.ErrorMessage,
			}
		}

		// Записываем событие в outbox
		outboxEvent := &broker.SagaStepResultEvent{
			Ref: broker.SagaStepRef{
				SagaId:      sagaInstanceID,
				StepName:    stepID,
				ServiceName: serviceName,
			},
			Worker: broker.WorkerInfo{
				InstanceId: "",
				Hostname:   "",
			},
			Status:     stepStatus,
			ResolvedAt: func() *time.Time { t := time.Now(); return &t }(),
			Result: map[string]any{
				"payment_id": paymentID.String(),
			},
			Error: errorInfo,
		}

		if err := s.outboxRepo.PushSagaStepResultEvent(ctx, outboxEvent); err != nil {
			captureErr = fmt.Errorf("push outbox event: %w", err)
			return captureErr
		}

		if !captureResp.Success {
			captureErr = fmt.Errorf("%s: %s", captureResp.ErrorCode, captureResp.ErrorMessage)
			return captureErr
		}

		result = &CaptureResult{
			OrderID:   orderID,
			PaymentID: paymentID,
		}

		return nil
	})

	if err != nil {
		return nil, captureErr
	}

	return result, nil
}
