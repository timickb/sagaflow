package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderEvent struct {
	OrderID uuid.UUID `json:"order_id"`
	Status  string    `json:"status"`
	Version uint32    `json:"version"`
	UserID  uuid.UUID `json:"user_id"`
}

type FctOrder struct {
	OrderID   uuid.UUID `ch:"order_id"`
	UserID    uuid.UUID `ch:"user_id"`
	Status    string    `ch:"status"`
	Version   uint32    `ch:"version"`
	CreatedAt time.Time `ch:"created_at"`
	UpdatedAt time.Time `ch:"updated_at"`
	LoadedAt  time.Time `ch:"loaded_at"`
}
