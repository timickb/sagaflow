package domain

import (
	"time"

	"github.com/google/uuid"
)

type Warehouse struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code      string    `gorm:"uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (Warehouse) TableName() string {
	return "warehouses"
}

type Product struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SKU       string    `gorm:"uniqueIndex;not null" json:"sku"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (Product) TableName() string {
	return "products"
}

type Balance struct {
	WarehouseID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"warehouse_id"`
	ProductID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"product_id"`
	QuantityTotal    int       `gorm:"not null" json:"quantity_total"`
	QuantityReserved int       `gorm:"not null;default:0" json:"quantity_reserved"`
	Version          int64     `gorm:"not null" json:"version"`
	UpdatedAt        time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (Balance) TableName() string {
	return "balances"
}

type MovementType string

const (
	MovementTypeInbound  MovementType = "INBOUND"
	MovementTypeOutbound MovementType = "OUTBOUND"
	MovementTypeReserve  MovementType = "RESERVE"
	MovementTypeRelease  MovementType = "RELEASE"
)

type Movement struct {
	MovementID         uuid.UUID    `gorm:"type:uuid;primaryKey" json:"movement_id"`
	WarehouseID        uuid.UUID    `gorm:"type:uuid;not null" json:"warehouse_id"`
	ProductID          uuid.UUID    `gorm:"type:uuid;not null" json:"product_id"`
	MovementType       MovementType `gorm:"not null" json:"movement_type"`
	Quantity           int          `gorm:"not null" json:"quantity"`
	BusinessRef        string       `gorm:"not null" json:"business_ref"` // order_id / shipment_id
	ScenarioInstanceID uuid.UUID    `gorm:"type:uuid;not null" json:"scenario_instance_id"`
	StepID             string       `gorm:"not null" json:"step_id"`
	CreatedAt          time.Time    `gorm:"not null;default:now()" json:"created_at"`
}

func (Movement) TableName() string {
	return "movements"
}

type ReservationStatus string

const (
	ReservationStatusActive    ReservationStatus = "ACTIVE"
	ReservationStatusConfirmed ReservationStatus = "CONFIRMED"
	ReservationStatusCancelled ReservationStatus = "CANCELLED"
)

type Reservation struct {
	ID                 uuid.UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID            uuid.UUID         `gorm:"not null" json:"order_id"`
	WarehouseID        uuid.UUID         `gorm:"type:uuid;not null" json:"warehouse_id"`
	ProductID          uuid.UUID         `gorm:"type:uuid;not null" json:"product_id"`
	Quantity           int               `gorm:"not null" json:"quantity"`
	Status             ReservationStatus `gorm:"not null" json:"status"`
	ScenarioInstanceID uuid.UUID         `gorm:"type:uuid;not null" json:"scenario_instance_id"`
	CreatedAt          time.Time         `gorm:"not null;default:now()" json:"created_at"`
}

func (Reservation) TableName() string {
	return "reservations"
}

// OutboxEvent - паттерн Outbox для атомарной публикации в Kafka
type OutboxEventStatus string

const (
	OutboxEventStatusPending   OutboxEventStatus = "PENDING"
	OutboxEventStatusPublished OutboxEventStatus = "PUBLISHED"
	OutboxEventStatusFailed    OutboxEventStatus = "FAILED"
)

type OutboxEvent struct {
	ID          uuid.UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	Topic       string            `gorm:"not null;index" json:"topic"`
	Key         string            `gorm:"not null" json:"key"` // saga_id для партиционирования
	Payload     string            `gorm:"type:jsonb;not null" json:"payload"`
	Status      OutboxEventStatus `gorm:"not null;default:'PENDING'" json:"status"`
	Attempts    int               `gorm:"not null;default:0" json:"attempts"`
	CreatedAt   time.Time         `gorm:"not null;default:now()" json:"created_at"`
	PublishedAt *time.Time        `gorm:"jsonb" json:"published_at"`
}

func (OutboxEvent) TableName() string {
	return "outbox_events"
}
