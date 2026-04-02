package domain

import (
	"github.com/google/uuid"
)

type StepUpdateDto struct {
	InstanceId uuid.UUID
	StepName   string
	Status     *StepStatus
	ErrorData  InstanceContext
}
