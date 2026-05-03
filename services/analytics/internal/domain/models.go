package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderEvent struct {
	OrderId     uuid.UUID `json:"order_id,omitempty"`
	UserId      uuid.UUID `json:"user_id,omitempty"`
	Status      string    `json:"status,omitempty"`
	Currency    string    `json:"currency,omitempty"`
	TotalAmount float64   `json:"total_amount,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Version     uint32    `json:"version,omitempty"`
}

type FctOrder struct {
	OrderID     uuid.UUID `ch:"order_id"`
	UserID      uuid.UUID `ch:"user_id"`
	Status      string    `ch:"status"`
	Version     uint32    `ch:"version"`
	TotalAmount float64   `ch:"total_amount"`
	Currnecy    string    `ch:"currnecy"`
	CreatedAt   time.Time `ch:"created_at"`
	UpdatedAt   time.Time `ch:"updated_at"`
	LoadedAt    time.Time `ch:"loaded_at"`
}
