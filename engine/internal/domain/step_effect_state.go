package domain

// StepEffectState - был ли реально зафиксирован результат шага?
type StepEffectState string

const (
	// StepEffectStateNone - локальная транзакция не закоммитилась, можно делать ретрай
	StepEffectStateNone StepEffectState = "NONE"
	// StepEffectStateCommitted - локальная транзакция точно зафиксировалась, ретраить нельзя
	StepEffectStateCommitted StepEffectState = "COMMITTED"
	// StepEffectStateUnknown - нельзя ретраить сразу, сначала нужно сделать сверку
	StepEffectStateUnknown StepEffectState = "UNKNOWN"
)
