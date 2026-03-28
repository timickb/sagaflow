package domain

// InstanceStartDto - данные для запуска экемпляра саги
type InstanceStartDto struct {
	// SagaName - название (= идентификатор) саги
	SagaName string
	// SagaVersion - версия саги
	SagaVersion string
	// InitialContext - входные параметры инстанса
	InitialContext InstanceContext

	// IdempotencyKey - опциональный ключ идемпотентности
	IdempotencyKey *string
	// CorrelationId - опциональный ключ контекста запроса
	CorrelationId *string
}
