package domain

import (
	"encoding/json"

	"github.com/google/uuid"
)

type StepCreateDto struct {
	InstanceId uuid.UUID
	StepName   string
	StepOrder  int
	InputData  json.RawMessage
}
