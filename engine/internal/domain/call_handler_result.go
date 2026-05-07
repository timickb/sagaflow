package domain

type CallHandlerResultStatus string

type CallHandlerResult struct {
	Status    CallHandlerResultStatus
	ErrorData *string
}

const (
	// CallHandlerResultSuccess - сервис принял шаг в обработку
	CallHandlerResultSuccess CallHandlerResultStatus = "SUCCESS"
	// CallHandlerResultUnprocessable - шаг не может быть обработан сервисом (в принципе)
	CallHandlerResultUnprocessable CallHandlerResultStatus = "UNPROCESSABLE"
	// CallHandlerResultHandlerNotFound - запрашиваемый сервис не зарегистрирован
	CallHandlerResultHandlerNotFound CallHandlerResultStatus = "HANDLER_NOT_FOUND"
)
