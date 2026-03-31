package dbstruct

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
)

type DBSagaStep struct {
	SagaId           uuid.UUID
	StepName         string
	StepOrder        int
	Status           domain.StepStatus
	Attempt          int
	WorkerInstanceId *string

	InputData  json.RawMessage
	OutputData json.RawMessage
	ErrorData  *json.RawMessage

	EffectState domain.StepEffectState

	StartedAt  *time.Time
	FinishedAt *time.Time
	UpdatedAt  time.Time
}

func (s *DBSagaStep) TableName() string {
	return "saga_step"
}

func (s *DBSagaStep) ToDomain() *domain.StepView {
	var errorData domain.InstanceContext
	if s.ErrorData != nil {
		errorData = domain.NewJsonInstanceContext(*s.ErrorData)
	}
	return &domain.StepView{
		Name:             s.StepName,
		Order:            s.StepOrder,
		Status:           s.Status,
		EffectState:      s.EffectState,
		Attempt:          s.Attempt,
		WorkerInstanceId: s.WorkerInstanceId,
		InputData:        domain.NewJsonInstanceContext(s.InputData),
		OutputData:       domain.NewJsonInstanceContext(s.OutputData),
		ErrorData:        errorData,
		StartedAt:        s.StartedAt,
		FinishedAt:       s.FinishedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}
