package domain

// InstanceStatus - статус экземпляра саги
type InstanceStatus string

const (
	InstanceStatusPending      InstanceStatus = "PENDING"
	InstanceStatusRunning      InstanceStatus = "RUNNING"
	InstanceStatusCompleted    InstanceStatus = "COMPLETED"
	InstanceStatusFailed       InstanceStatus = "FAILED"
	InstanceStatusCompensating InstanceStatus = "COMPENSATING"
	InstanceStatusCompensated  InstanceStatus = "COMPENSATED"
	InstanceStatusInconsistent InstanceStatus = "INCONSISTENT"
)
