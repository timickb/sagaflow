package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PaymentProvider - интерфейс платежной системы
type PaymentProvider interface {
	// Capture выполняет списание средств
	Capture(ctx context.Context, req *CaptureRequest) (*CaptureResponse, error)
}

type CaptureRequest struct {
	PaymentID     uuid.UUID
	Amount        float64
	Currency      string
	PaymentMethod string
	OrderID       uuid.UUID
}

type CaptureResponse struct {
	Success           bool
	ProviderPaymentID string
	ErrorCode         string
	ErrorMessage      string
}

// StubPaymentProvider - заглушка для платежной системы
type StubPaymentProvider struct{}

func NewStubPaymentProvider() *StubPaymentProvider {
	return &StubPaymentProvider{}
}

func (p *StubPaymentProvider) Capture(ctx context.Context, req *CaptureRequest) (*CaptureResponse, error) {
	// Имитация задержки платежной системы
	time.Sleep(100 * time.Millisecond)

	// Генерируем ID транзакции
	providerPaymentID := uuid.New().String()

	// Заглушка: считаем успешным, если сумма > 0
	if req.Amount <= 0 {
		return &CaptureResponse{
			Success:      false,
			ErrorCode:    "INVALID_AMOUNT",
			ErrorMessage: "Amount must be greater than 0",
		}, nil
	}

	// Заглушка: симуляция ошибок для определенных сумм
	if req.Amount > 1000000 {
		return &CaptureResponse{
			Success:           false,
			ProviderPaymentID: providerPaymentID,
			ErrorCode:         "AMOUNT_LIMIT_EXCEEDED",
			ErrorMessage:      fmt.Sprintf("Amount %.2f exceeds maximum allowed", req.Amount),
		}, nil
	}

	return &CaptureResponse{
		Success:           true,
		ProviderPaymentID: providerPaymentID,
	}, nil
}

// Ensure StubPaymentProvider implements PaymentProvider
var _ PaymentProvider = (*StubPaymentProvider)(nil)
