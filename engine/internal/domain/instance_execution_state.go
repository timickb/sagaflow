package domain

// InstanceExecutionState - техническое состояние инстанса для раннера
type InstanceExecutionState string

const (
	// InstanceExecutionStateRunnable - можно выполнять следующий шаг
	InstanceExecutionStateRunnable = "RUNNABLE"
	// InstanceExecutionStateWaitEvent - инстанс ожидает события от сервиса-обработчика
	InstanceExecutionStateWaitEvent = "WAIT_EVENT"
)
