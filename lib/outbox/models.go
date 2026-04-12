package outbox

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/lib/broker"
)

type DBEvent struct {
	Id            uuid.UUID       `gorm:"id"`
	AggregateType string          `gorm:"aggregatetype"`
	AggregateId   string          `gorm:"aggregateid"`
	Payload       json.RawMessage `gorm:"payload"`
	CreatedAt     time.Time       `gorm:"created_at"`
}

// NewEventDtoFromStepResultEvent - создать запись для таблицы outbox_events из структуры SagaStepResultEvent
func NewEventDtoFromStepResultEvent(event *broker.SagaStepResultEvent) (*DBEvent, error) {
	eventMarshaled, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}
	return &DBEvent{
		Id:            uuid.New(),
		AggregateType: "saga.step.result",
		AggregateId:   "saga.step.result",
		Payload:       eventMarshaled,
		CreatedAt:     time.Now(),
	}, nil
}

func (e *DBEvent) TableName() string {
	return "outbox_events"
}
