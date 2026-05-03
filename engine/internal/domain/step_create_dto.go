package domain

import (
	"github.com/google/uuid"
)

type StepCreateDto struct {
	InstanceId uuid.UUID
	StepName   string
	StepOrder  int
	// Status - по умолчанию PENDING
	Status    *StepStatus
	InputData InstanceContext
}
