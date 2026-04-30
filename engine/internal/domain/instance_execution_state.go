package domain

// InstanceExecutionState - техническое состояние инстанса для раннера
type InstanceExecutionState string

const (
	// InstanceExecutionStateRunnable - можно выполнять следующий шаг
	InstanceExecutionStateRunnable InstanceExecutionState = "RUNNABLE"
	// InstanceExecutionStateWaitingEvent - инстанс ожидает события от сервиса-обработчика
	InstanceExecutionStateWaitingEvent InstanceExecutionState = "WAITING_EVENT"
)
