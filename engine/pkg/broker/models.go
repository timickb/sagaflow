package broker

import "time"

type SagaStepResultStatus string

const (
	SagaStepStatusCommitted SagaStepResultStatus = "COMMITTED"
	SagaStepStatusFailed    SagaStepResultStatus = "FAILED"
	SagaStepStatusRejected  SagaStepResultStatus = "REJECTED"
)

// SagaStepRef - шаг саги
type SagaStepRef struct {
	SagaId      string `json:"saga_id"`
	StepId      string `json:"step_id"`
	StepName    string `json:"step_name"`
	ServiceName string `json:"service_name"`
}

// WorkerInfo - инфо о воркере, отправившем результат
type WorkerInfo struct {
	InstanceId string `json:"instance_id"`
	Hostname   string `json:"hostname"`
}

// ErrorInfo - структура ошибки выполнения шага
type ErrorInfo struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Retriable bool                   `json:"retriable"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// SagaStepResultEvent - модель события из топика
type SagaStepResultEvent struct {
	Ref    SagaStepRef          `json:"ref"`
	Worker WorkerInfo           `json:"worker"`
	Status SagaStepResultStatus `json:"status"`

	CommittedAt *time.Time `json:"committed_at,omitempty"`
	FailedAt    *time.Time `json:"failed_at,omitempty"`
	RejectedAt  *time.Time `json:"rejected_at,omitempty"`

	Result map[string]interface{} `json:"result,omitempty"`
	Error  *ErrorInfo             `json:"error,omitempty"`
	Effect string                 `json:"effect,omitempty"`
}
