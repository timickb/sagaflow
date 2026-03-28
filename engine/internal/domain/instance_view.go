package domain

import (
	"time"

	"github.com/google/uuid"
)

type InstanceView struct {
	SagaId      uuid.UUID
	SagaName    string
	SagaVersion string

	Status InstanceStatus

	InitialContext InstanceContext
	RuntimeContext InstanceContext
	ContextVersion int64

	IdempotencyKey string

	CurrentStepName  *string
	CorrelationId    *string
	LastErrorCode    *string
	LastErrorMessage *string

	StartedAt  time.Time
	UpdatedAt  time.Time
	FinishedAt *time.Time
}
