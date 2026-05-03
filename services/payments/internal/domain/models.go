package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusRefunded  OrderStatus = "REFUNDED"
)

type Order struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID       `gorm:"type:uuid;not null" json:"user_id"`
	Status      OrderStatus     `gorm:"type:text;not null;default:'PENDING'" json:"status"`
	TotalAmount float64         `gorm:"type:numeric(12,2);not null" json:"total_amount"`
	Currency    string          `gorm:"type:text;not null;default:'USD'" json:"currency"`
	Version     int             `gorm:"type:int;not null" json:"version"`
	Details     json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"details"`
	CreatedAt   time.Time       `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"not null;default:now()" json:"updated_at"`
}

func (Order) TableName() string {
	return "orders"
}

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "PENDING"
	PaymentStatusAuthorized PaymentStatus = "AUTHORIZED"
	PaymentStatusCaptured   PaymentStatus = "CAPTURED"
	PaymentStatusFailed     PaymentStatus = "FAILED"
	PaymentStatusRefunded   PaymentStatus = "REFUNDED"
)

type Payment struct {
	ID                uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID           uuid.UUID     `gorm:"type:uuid;not null" json:"order_id"`
	Amount            float64       `gorm:"type:numeric(12,2);not null" json:"amount"`
	Currency          string        `gorm:"type:text;not null;default:'RUB'" json:"currency"`
	PaymentMethod     string        `gorm:"type:text;not null" json:"payment_method"`
	PaymentProvider   string        `gorm:"type:text;not null" json:"payment_provider"`
	ProviderPaymentID *string       `gorm:"type:text" json:"provider_payment_id,omitempty"`
	Status            PaymentStatus `gorm:"type:text;not null;default:'PENDING'" json:"status"`
	CapturedAt        *time.Time    `gorm:"type:timestamptz" json:"captured_at,omitempty"`
	RefundedAt        *time.Time    `gorm:"type:timestamptz" json:"refunded_at,omitempty"`
	CreatedAt         time.Time     `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt         time.Time     `gorm:"not null;default:now()" json:"updated_at"`
}

func (Payment) TableName() string {
	return "payments"
}

type OrderAnalyticsEvent struct {
	OrderId     uuid.UUID `json:"order_id,omitempty"`
	UserId      uuid.UUID `json:"user_id,omitempty"`
	Status      string    `json:"status,omitempty"`
	Currency    string    `json:"currency,omitempty"`
	TotalAmount float64   `json:"total_amount,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Version     int       `json:"version,omitempty"`
}
