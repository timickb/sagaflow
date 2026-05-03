package outbox

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/lib/broker"
)

const (
	stepResultTopic = "saga.step.result"
)

type DBEvent struct {
	Id            uuid.UUID
	Aggregatetype string
	Aggregateid   string
	Type          string
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
		Aggregatetype: stepResultTopic,
		Aggregateid:   stepResultTopic,
		Type:          stepResultTopic,
		Payload:       eventMarshaled,
		CreatedAt:     time.Now(),
	}, nil
}

func (e *DBEvent) TableName() string {
	return "outbox_events"
}
