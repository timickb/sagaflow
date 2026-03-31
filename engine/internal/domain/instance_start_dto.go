package domain

// InstanceStartDto - данные для запуска экемпляра саги
type InstanceStartDto struct {
	// SagaName - название (= идентификатор) саги
	SagaName string
	// SagaVersion - версия саги
	SagaVersion int
	// InitialContext - входные параметры инстанса
	InitialContext InstanceContext

	// IdempotencyKey - опциональный ключ идемпотентности
	// по умолчанию = random uuid
	IdempotencyKey *string
	// CorrelationId - опциональный ключ контекста запроса
	CorrelationId *string
}
