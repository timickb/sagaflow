package domain

import "github.com/timickb/sagaflow/lib/utils"

type (
	// StepKind - тип шага
	StepKind string
	// StepOutcome - результат шага
	StepOutcome string
	// VerifierType - тип способа сверки данных
	VerifierType string
	// RetryBackoffType - тип ретрая
	RetryBackoffType string
	// SagaResult - финальный статус саги
	SagaResult string
)

const (
	// StepKindAction - прямой шаг саги
	StepKindAction StepKind = "action"
	// StepKindVerify - шаг сверки данных во внешних системах
	StepKindVerify StepKind = "verify"
	// StepKindReconcile - шаг восстановления консистентности данных во внешности системах
	StepKindReconcile StepKind = "reconcile"
	// StepKindCompensate - шаг компенсации изменений для action
	StepKindCompensate StepKind = "compensate"
	// StepKindTerminal - терминальное (финальное) состояние саги
	StepKindTerminal StepKind = "terminal"

	// OutcomeCommitted - шаг успешно выполнен, результат зафиксирован
	OutcomeCommitted StepOutcome = "committed"
	// OutcomeRejected - выполнение шага отклонено
	OutcomeRejected StepOutcome = "rejected"
	// OutcomeFailed - выполнение шага завершено с ошибкой
	OutcomeFailed StepOutcome = "failed"
	// OutcomeTimeout - выполнение шага не уложилось в заданный таймаут
	OutcomeTimeout StepOutcome = "timeout"
	// OutcomeMatched - проверка успешно пройдена (для verify шагов)
	OutcomeMatched StepOutcome = "matched"
	// OutcomeUnmatched - обнаружено расхождение в данных (для verify шагов)
	OutcomeUnmatched StepOutcome = "unmatched"

	VerifierTypeSql VerifierType = "sql"
	VerifierTypeApi VerifierType = "api"

	RetryBackoffTypeFixed       RetryBackoffType = "fixed"
	RetryBackoffTypeExponential RetryBackoffType = "exponential"

	// SagaResultCompleted - сага завершена успешно
	SagaResultCompleted SagaResult = SagaResult(InstanceStatusCompleted)
	// SagaResultFailed - не удалось завершить сагу
	SagaResultFailed SagaResult = SagaResult(InstanceStatusFailed)
	// SagaResultCompensated - сага завершена с компенсацией шагов
	SagaResultCompensated SagaResult = SagaResult(InstanceStatusCompensated)
	// SagaResultInconsistent - не удалось подтвердить end-to-end консистентность
	SagaResultInconsistent SagaResult = SagaResult(InstanceStatusInconsistent)
)

func (o StepOutcome) IsValid() bool {
	switch o {
	case OutcomeCommitted, OutcomeRejected, OutcomeFailed, OutcomeTimeout, OutcomeMatched, OutcomeUnmatched:
		return true
	default:
		return false
	}
}

func (r SagaResult) IsValid() bool {
	switch r {
	case SagaResultCompleted, SagaResultFailed, SagaResultCompensated, SagaResultInconsistent:
		return true
	default:
		return false
	}
}

func (r VerifierType) IsValid() bool {
	switch r {
	case VerifierTypeSql, VerifierTypeApi:
		return true
	default:
		return false
	}
}

func (r RetryBackoffType) IsValid() bool {
	switch r {
	case RetryBackoffTypeFixed, RetryBackoffTypeExponential:
		return true
	default:
		return false
	}
}

// ToInstanceStatus - в какой статус нужно перевести экземпляр, если тип следующего шага = s.Kind?
func (s DefinitionStep) ToInstanceStatus(currentStatus InstanceStatus) (newStatus *InstanceStatus) {
	switch s.Kind {
	case StepKindAction:
		return utils.Ptr(InstanceStatusRunning)
	case StepKindCompensate:
		return utils.Ptr(InstanceStatusCompensating)
	case StepKindVerify, StepKindReconcile:
		return utils.Ptr(InstanceStatusVerifying)
	case StepKindTerminal:
		switch currentStatus {
		case InstanceStatusCompensating:
			return utils.Ptr(InstanceStatusCompensated)
		default:
			return utils.Ptr(InstanceStatus(*s.Result))
		}
	default:
		return nil
	}
}
