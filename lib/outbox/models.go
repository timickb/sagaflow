package outbox

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/lib/broker"
)

const (
	stepResultTopic    = "saga.step.result"
	eventAggregateType = "SagaInstance"
	eventType          = "SagaStepCommitted"
)

type DBEvent struct {
	Id            uuid.UUID
	AggregateType string
	AggregateId   string
	EventType     string
	Payload       json.RawMessage
	CreatedAt     time.Time
}

// NewEventDtoFromStepResultEvent - создать запись для таблицы outbox_events из структуры SagaStepResultEvent
func NewEventDtoFromStepResultEvent(event *broker.SagaStepResultEvent) (*DBEvent, error) {
	eventMarshaled, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}
	return &DBEvent{
		Id:            uuid.New(),
		AggregateType: eventAggregateType,
		AggregateId:   event.Ref.SagaId.String(),
		EventType:     eventType,
		Payload:       eventMarshaled,
		CreatedAt:     time.Now(),
	}, nil
}

func (e *DBEvent) TableName() string {
	return "saga_outbox_events"
}
