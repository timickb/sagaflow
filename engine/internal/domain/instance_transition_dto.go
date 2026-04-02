package domain

import (
	"time"

	"github.com/google/uuid"
)

// InstanceTransitionDto - данные для перевода экземпляра на следующий шаг
type InstanceTransitionDto struct {
	Id             uuid.UUID
	RuntimeContext InstanceContext
	NextStepName   string
	Status         *InstanceStatus
	ExecutionState *InstanceExecutionState
	ErrCode        *string
	ErrMessage     *string

	NextExecutionAt *time.Time
}
