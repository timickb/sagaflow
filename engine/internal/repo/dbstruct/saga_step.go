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

// NewSagaStep - создать шаг экземпляра, ожидающий выполнения
func NewSagaStep(dto *domain.StepCreateDto) *DBSagaStep {
	return &DBSagaStep{
		SagaId:    dto.InstanceId,
		StepName:  dto.StepName,
		StepOrder: dto.StepOrder,
		Status:    domain.StepStatusPending,
		Attempt:   1,
		InputData: dto.InputData,
		UpdatedAt: time.Now(),
	}
}

func (s *DBSagaStep) TableName() string {
	return "saga_step"
}

func (s *DBSagaStep) ToDomain() *domain.StepView {
	var errorData domain.InstanceContext
	if s.ErrorData != nil {
		errorData = domain.NewJsonInstanceContextFromRaw(*s.ErrorData)
	}
	return &domain.StepView{
		Name:             s.StepName,
		Order:            s.StepOrder,
		Status:           s.Status,
		EffectState:      s.EffectState,
		Attempt:          s.Attempt,
		WorkerInstanceId: s.WorkerInstanceId,
		InputData:        domain.NewJsonInstanceContextFromRaw(s.InputData),
		OutputData:       domain.NewJsonInstanceContextFromRaw(s.OutputData),
		ErrorData:        errorData,
		StartedAt:        s.StartedAt,
		FinishedAt:       s.FinishedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}

func NewSagaStepUpdatesMap(dto *domain.StepUpdateDto) map[string]interface{} {
	now := time.Now()
	result := map[string]interface{}{
		"updated_at": now,
	}
	return result
}
