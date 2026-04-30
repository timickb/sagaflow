package domain

// InstanceTerminateDto - данные для перевода экземпляра в терминальный статус
type InstanceTerminateDto struct {
	Status     InstanceStatus
	ErrCode    *string
	ErrMessage *string
}
