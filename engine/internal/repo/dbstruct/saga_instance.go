package dbstruct

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
)

// DBSagaInstance - проекция таблицы saga_instance
type DBSagaInstance struct {
	SagaId      uuid.UUID
	SagaName    string
	SagaVersion int

	Status         domain.InstanceStatus
	ExecutionState domain.InstanceExecutionState

	CurrentStepName *string
	IdempotencyKey  string
	CorrelationId   *string

	InitialContext json.RawMessage
	RuntimeContext json.RawMessage

	StartedAt  time.Time
	UpdatedAt  time.Time
	FinishedAt *time.Time

	LastErrorCode    *string
	LastErrorMessage *string

	ContextVersion int64

	NextExecutionAt time.Time
	LockedTill      *time.Time
	LockedBy        *string
}

// NewSagaInstance - создать экземпляр саги, ожидающий начала выполнения
func NewSagaInstance(id uuid.UUID, dto *domain.InstanceStartDto) *DBSagaInstance {
	var idempotencyKey string
	if dto.IdempotencyKey == nil {
		idempotencyKey = uuid.New().String()
	}
	now := time.Now()
	return &DBSagaInstance{
		SagaId:          id,
		SagaName:        dto.SagaName,
		SagaVersion:     dto.SagaVersion,
		Status:          domain.InstanceStatusPending,
		ExecutionState:  domain.InstanceExecutionStateRunnable,
		CurrentStepName: &dto.StartStepName,
		IdempotencyKey:  idempotencyKey,
		CorrelationId:   dto.CorrelationId,
		InitialContext:  dto.InitialContext.GetRaw(),
		RuntimeContext:  domain.EmptyInstanceContext.GetRaw(),
		StartedAt:       now,
		UpdatedAt:       now,
	}
}

func (si *DBSagaInstance) TableName() string {
	return "saga_instance"
}

func (si *DBSagaInstance) ToDomain() (*domain.InstanceView, error) {
	initialContext, err := domain.NewJsonInstanceContextFromRaw(si.InitialContext)
	if err != nil {
		return nil, fmt.Errorf("take saga instance initial context: %w", err)
	}
	runtimeContext, err := domain.NewJsonInstanceContextFromRaw(si.RuntimeContext)
	if err != nil {
		return nil, fmt.Errorf("take saga instance runtime context: %w", err)
	}
	return &domain.InstanceView{
		SagaId:           si.SagaId,
		SagaName:         si.SagaName,
		SagaVersion:      si.SagaVersion,
		Status:           si.Status,
		ExecutionState:   si.ExecutionState,
		InitialContext:   initialContext,
		RuntimeContext:   runtimeContext,
		ContextVersion:   si.ContextVersion,
		IdempotencyKey:   si.IdempotencyKey,
		CurrentStepName:  si.CurrentStepName,
		CorrelationId:    si.CorrelationId,
		LastErrorCode:    si.LastErrorCode,
		LastErrorMessage: si.LastErrorMessage,
		StartedAt:        si.StartedAt,
		UpdatedAt:        si.UpdatedAt,
		FinishedAt:       si.FinishedAt,
	}, nil
}

func NewSagaInstanceMakeTransitionUpdateMap(dto *domain.InstanceTransitionDto) map[string]interface{} {
	now := time.Now()
	result := map[string]interface{}{
		"updated_at":        now,
		"locked_till":       (*time.Time)(nil),
		"locked_by":         (*string)(nil),
		"current_step_name": dto.NextStepName,
	}
	if dto.Status != nil {
		result["status"] = *dto.Status
		if dto.Status.IsTerminal() {
			result["finished_at"] = now
		}
	}
	if dto.ExecutionState != nil {
		result["execution_state"] = *dto.ExecutionState
	}
	if dto.RuntimeContext != nil {
		result["runtime_context"] = dto.RuntimeContext.GetRaw()
	}
	if dto.ErrCode != nil {
		result["last_error_code"] = *dto.ErrCode
	}
	if dto.ErrMessage != nil {
		result["last_error_message"] = *dto.ErrMessage
	}
	if dto.NextExecutionAt != nil {
		result["next_execution_at"] = *dto.NextExecutionAt
	}
	return result
}
