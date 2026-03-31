package domain

// InstanceStatus - статус экземпляра саги.
type InstanceStatus string

const (
	// InstanceStatusPending - экземпляр ожидает начала выполнения.
	InstanceStatusPending InstanceStatus = "PENDING"
	// InstanceStatusRunning - идет выполнение основной части саги.
	InstanceStatusRunning InstanceStatus = "RUNNING"
	// InstanceStatusCompleted - экземпляр успешно завершен (терминальный статус).
	InstanceStatusCompleted InstanceStatus = "COMPLETED"
	// InstanceStatusFailed - экземпляр завершен с ошибкой (терминальный статус).
	InstanceStatusFailed InstanceStatus = "FAILED"
	// InstanceStatusVerifying - идет выполнение сверок данных в системах.
	InstanceStatusVerifying InstanceStatus = "VERIFYING"
	// InstanceStatusCompensating - идет выполнение компенсационных шагов.
	InstanceStatusCompensating InstanceStatus = "COMPENSATING"
	// InstanceStatusCompensated - экземпляр завершен с компенсацией шагов (терминальный статус).
	InstanceStatusCompensated InstanceStatus = "COMPENSATED"
	// InstanceStatusInconsistent - экземпляр завершен, но не удалось обеспечить консистентность.
	// Требует ручного разбора.
	InstanceStatusInconsistent InstanceStatus = "INCONSISTENT"
)
