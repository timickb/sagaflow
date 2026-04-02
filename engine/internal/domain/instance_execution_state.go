package domain

// InstanceExecutionState - техническое состояние инстанса для раннера
type InstanceExecutionState string

const (
	// InstanceExecutionStateRunnable - можно выполнять следующий шаг
	InstanceExecutionStateRunnable InstanceExecutionState = "RUNNABLE"
	// InstanceExecutionStateWaitEvent - инстанс ожидает события от сервиса-обработчика
	InstanceExecutionStateWaitEvent InstanceExecutionState = "WAIT_EVENT"
)
