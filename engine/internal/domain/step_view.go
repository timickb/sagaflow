package domain

import (
	"time"

	"github.com/google/uuid"
)

// StepView - представление конкретного шага экземпляра саги
type StepView struct {
	// SagaId - ID экземпляра
	SagaId uuid.UUID
	// Name - название шага в DSL (= идентификатор)
	Name string
	// Order - порядковый номер шага
	Order int
	// Status - статус шага
	Status StepStatus
	// EffectState - статус фиксации результата обработчика шага
	EffectState StepEffectState
	// Attempt - номер попытки выполнения
	Attempt          int
	WorkerInstanceId *string

	// InputData - какие данные пришли на вход в обработчик
	InputData InstanceContext
	// OutputData - какие данные порождены обработчиком в качестве результата
	OutputData InstanceContext
	// ErrorData - данные об ошибках, сгенерированные обработчиком
	ErrorData InstanceContext

	// StartedAt - когда началось выполнение шага
	StartedAt *time.Time
	// FinishedAt - когда завершилось выполнение шага
	FinishedAt *time.Time
	UpdatedAt  time.Time
}
