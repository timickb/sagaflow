package dbstruct

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
)

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
		CurrentStepName: nil,
		IdempotencyKey:  idempotencyKey,
		CorrelationId:   dto.CorrelationId,
		InitialContext:  dto.InitialContext.GetRaw(),
		StartedAt:       now,
		UpdatedAt:       now,
	}
}

func (si *DBSagaInstance) TableName() string {
	return "saga_instance"
}

func (si *DBSagaInstance) ToDomain() *domain.InstanceView {
	return &domain.InstanceView{
		SagaId:           si.SagaId,
		SagaName:         si.SagaName,
		SagaVersion:      si.SagaVersion,
		Status:           si.Status,
		InitialContext:   domain.NewJsonInstanceContext(si.InitialContext),
		RuntimeContext:   domain.NewJsonInstanceContext(si.RuntimeContext),
		ContextVersion:   si.ContextVersion,
		IdempotencyKey:   si.IdempotencyKey,
		CurrentStepName:  si.CurrentStepName,
		CorrelationId:    si.CorrelationId,
		LastErrorCode:    si.LastErrorCode,
		LastErrorMessage: si.LastErrorMessage,
		StartedAt:        si.StartedAt,
		UpdatedAt:        si.UpdatedAt,
		FinishedAt:       si.FinishedAt,
	}
}
