package broker

import (
	"time"

	"github.com/google/uuid"
)

type SagaStepResultStatus string

const (
	SagaStepStatusCommitted SagaStepResultStatus = "COMMITTED"
	SagaStepStatusFailed    SagaStepResultStatus = "FAILED"
	SagaStepStatusRejected  SagaStepResultStatus = "REJECTED"
)

// SagaStepRef - шаг саги
type SagaStepRef struct {
	SagaId      uuid.UUID `json:"saga_id"`
	StepName    string    `json:"step_name"`
	ServiceName string    `json:"service_name"`
}

// WorkerInfo - инфо о воркере, отправившем результат
type WorkerInfo struct {
	InstanceId string `json:"instance_id"`
	Hostname   string `json:"hostname"`
}

// ErrorInfo - структура ошибки выполнения шага
type ErrorInfo struct {
	// Code - код ошибки
	Code string `json:"code"`
	// Message - подробное сообщение
	Message string `json:"message"`
	// Retriable - если false, то ретраить нельзя, нужно сразу делать компенсацию
	Retriable bool                   `json:"retriable"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// SagaStepResultEvent - модель события из топика
type SagaStepResultEvent struct {
	Ref    SagaStepRef          `json:"ref"`
	Worker WorkerInfo           `json:"worker"`
	Status SagaStepResultStatus `json:"status"`

	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// Result - данные, которые нужно записать в runtime context
	Result map[string]interface{} `json:"result,omitempty"`
	Error  *ErrorInfo             `json:"error,omitempty"`
}
