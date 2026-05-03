package service

import (
	"github.com/google/uuid"
	"github.com/timickb/sagaflow/services/payments/internal/domain"
	"github.com/timickb/sagaflow/services/payments/internal/payment"
	"github.com/timickb/sagaflow/services/payments/internal/repository"
)

const serviceName = "payments"

type CapturePayload struct {
	Items   []interface{} `json:"items"`
	UserID  string        `json:"user_id"`
	OrderID string        `json:"order_id"`
	Amount  int           `json:"amount"`
}

type CaptureResult struct {
	OrderID   uuid.UUID
	PaymentID uuid.UUID
}

type PaymentsService struct {
	orderRepo   *repository.OrderRepository
	paymentRepo *repository.PaymentRepository
	outboxRepo  *repository.OutboxRepository
	transactor  domain.Transactor
	paymentProv payment.PaymentProvider
}

func NewPaymentsService(
	orderRepo *repository.OrderRepository,
	paymentRepo *repository.PaymentRepository,
	outboxRepo *repository.OutboxRepository,
	transactor domain.Transactor,
	paymentProv payment.PaymentProvider,
) *PaymentsService {
	return &PaymentsService{
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
		outboxRepo:  outboxRepo,
		transactor:  transactor,
		paymentProv: paymentProv,
	}
}
