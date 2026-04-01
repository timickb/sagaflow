package domain

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
	// OutcomePassed - проверка успешно пройдена (для verify шагов)
	OutcomePassed StepOutcome = "passed"
	// OutcomeVerificationFailed - обнаружено расхождение в данных (для verify шагов)
	OutcomeVerificationFailed StepOutcome = "verification_failed"

	VerifierTypeSql       VerifierType = "sql"
	VerifierTypeComposite VerifierType = "composite"
	VerifierTypeApi       VerifierType = "api"

	RetryBackoffTypeFixed       RetryBackoffType = "fixed"
	RetryBackoffTypeExponential RetryBackoffType = "exponential"

	// SagaResultCompleted - сага завершена успешно
	SagaResultCompleted SagaResult = "COMPLETED"
	// SagaResultFailed - не удалось завершить сагу
	SagaResultFailed SagaResult = "FAILED"
	// SagaResultCompensated - сага завершена с компенсацией шагов
	SagaResultCompensated SagaResult = "COMPENSATED"
	// SagaResultInconsistent - не удалось подтвердить end-to-end консистентность
	SagaResultInconsistent SagaResult = "INCONSISTENT"
)

func (o StepOutcome) IsValid() bool {
	switch o {
	case OutcomeCommitted, OutcomeRejected, OutcomeFailed, OutcomeTimeout, OutcomePassed, OutcomeVerificationFailed:
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
	case VerifierTypeSql, VerifierTypeComposite, VerifierTypeApi:
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
