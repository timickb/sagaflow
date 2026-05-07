package domain

import "time"

// InstanceStartDto - данные для запуска экемпляра саги
type InstanceStartDto struct {
	// SagaName - название (= идентификатор) саги
	SagaName string
	// SagaVersion - версия саги
	SagaVersion int
	// InitialContext - входные параметры инстанса
	InitialContext InstanceContext

	// StartStepName - название (= идентификатор) первого шага
	StartStepName string

	// IdempotencyKey - опциональный ключ идемпотентности
	// по умолчанию = random uuid
	IdempotencyKey *string
	// CorrelationId - опциональный ключ контекста запроса
	CorrelationId *string

	// StartAfter - время отложенного запуска
	StartAfter *time.Time
}
