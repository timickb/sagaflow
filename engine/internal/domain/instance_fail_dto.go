package domain

// InstanceFailDto - данные для перевода экземпляра в статус FAILED
type InstanceFailDto struct {
	ErrCode    string
	ErrMessage *string
}
