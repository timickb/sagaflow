package domain

// InstanceFailReason - причина перехода экземпляра в статус FAILED
type InstanceFailReason string

const (
	// InstanceFailReasonSagaNotFound - сага не найдена
	InstanceFailReasonSagaNotFound InstanceFailReason = "SAGA_NOT_FOUND"
	// InstanceFailReasonStepNotFound - ожидаемый шаг не задекларирован
	InstanceFailReasonStepNotFound InstanceFailReason = "STEP_NOT_FOUND"
	// InstanceFailReasonInconsistentStep - состояние текущего шага не сохранено в БД
	InstanceFailReasonInconsistentStep InstanceFailReason = "INCONSISTENT_STEP"
	// InstanceFailReasonUnknownEventStatus - обработчик вернул несуществующий статус
	InstanceFailReasonUnknownEventStatus InstanceFailReason = "UNKNOWN_EVENT_STATUS"
	// InstanceFailReasonRetriesExceeded - не удалось получить ответ от воркера после всех ретраев
	InstanceFailReasonRetriesExceeded InstanceFailReason = "RETRIES_EXCEEDED"
)
