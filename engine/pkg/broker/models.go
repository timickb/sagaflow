package broker

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
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
	Result map[string]any `json:"result,omitempty"`
	Error  *ErrorInfo     `json:"error,omitempty"`
}

func (e ErrorInfo) String() string {
	base := fmt.Sprintf("[%s] %s", e.Code, e.Message)

	if len(e.Details) == 0 {
		return base
	}

	keys := make([]string, 0, len(e.Details))
	for k := range e.Details {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, stringifyDetailValue(e.Details[k])))
	}

	return fmt.Sprintf("%s: %s", base, strings.Join(parts, ", "))
}

func stringifyDetailValue(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return fmt.Sprintf("%v", x)
		}
		return string(b)
	}
}
