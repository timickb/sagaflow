package dbstruct

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"gorm.io/gorm"
)

type DBSagaStep struct {
	SagaId           uuid.UUID
	StepName         string
	StepOrder        int
	Status           domain.StepStatus
	Attempt          int
	ReconcileCycles  int
	WorkerInstanceId *string

	InputData  json.RawMessage
	OutputData json.RawMessage
	ErrorData  *json.RawMessage

	EffectState *domain.StepEffectState

	StartedAt  *time.Time
	FinishedAt *time.Time
	UpdatedAt  time.Time
}

// NewSagaStep - создать шаг экземпляра, ожидающий выполнения
func NewSagaStep(dto *domain.StepCreateDto) *DBSagaStep {
	status := domain.StepStatusPending
	if dto.Status != nil {
		status = *dto.Status
	}
	return &DBSagaStep{
		SagaId:          dto.InstanceId,
		StepName:        dto.StepName,
		StepOrder:       dto.StepOrder,
		Status:          status,
		Attempt:         1,
		ReconcileCycles: 0,
		InputData:       dto.InputData.GetRaw(),
		OutputData:      domain.EmptyInstanceContext.GetRaw(),
		UpdatedAt:       time.Now(),
	}
}

func (s *DBSagaStep) TableName() string {
	return "saga_step"
}

func (s *DBSagaStep) ToDomain() (*domain.StepView, error) {
	var (
		errorData, inputData, outputData domain.InstanceContext
		err                              error
	)
	if s.ErrorData != nil {
		errorData, err = domain.NewJsonInstanceContextFromRaw(*s.ErrorData)
		if err != nil {
			return nil, fmt.Errorf("errorData unmarshalling error during step mapping: %w", err)
		}
	}
	inputData, err = domain.NewJsonInstanceContextFromRaw(s.InputData)
	if err != nil {
		return nil, fmt.Errorf("inputData unmarshalling error during step mapping: %w", err)
	}
	outputData, err = domain.NewJsonInstanceContextFromRaw(s.OutputData)
	if err != nil {
		return nil, fmt.Errorf("outputData unmarshalling error during step mapping: %w", err)
	}
	return &domain.StepView{
		SagaId:           s.SagaId,
		Name:             s.StepName,
		Order:            s.StepOrder,
		Status:           s.Status,
		EffectState:      s.EffectState,
		Attempt:          s.Attempt,
		ReconcileCycles:  s.ReconcileCycles,
		WorkerInstanceId: s.WorkerInstanceId,
		InputData:        inputData,
		OutputData:       outputData,
		ErrorData:        errorData,
		StartedAt:        s.StartedAt,
		FinishedAt:       s.FinishedAt,
		UpdatedAt:        s.UpdatedAt,
	}, nil
}

func NewSagaStepUpdatesMap(dto *domain.StepUpdateDto) map[string]interface{} {
	now := time.Now()
	result := map[string]interface{}{
		"updated_at": now,
	}
	if dto.Status != nil {
		result["status"] = *dto.Status
		if dto.Status.IsTerminal() {
			result["finished_at"] = now
		}
	}
	if dto.IncrementAttempt {
		result["attempt"] = gorm.Expr("attempt + 1")
	}
	if dto.ErrorData != nil {
		result["error_data"] = dto.ErrorData.GetRaw()
	}
	return result
}
