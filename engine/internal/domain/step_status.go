package domain

// StepStatus - статус шага инстанса
type StepStatus string

const (
	StepStatusPending      StepStatus = "PENDING"
	StepStatusRunning      StepStatus = "RUNNING"
	StepStatusCommitted    StepStatus = "COMMITTED"
	StepStatusFailed       StepStatus = "FAILED"
	StepStatusCompensating StepStatus = "COMPENSATING"
	StepStatusCompensated  StepStatus = "COMPENSATED"
	StepStatusVerifying    StepStatus = "VERIFYING"
	StepStatusVerified     StepStatus = "VERIFIED"
)

func (s StepStatus) IsTerminal() bool {
	switch s {
	case StepStatusCompensated, StepStatusVerified, StepStatusFailed, StepStatusCommitted:
		return true
	default:
		return false
	}
}
