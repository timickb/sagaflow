package domain

import (
	"time"
)

type SagaDefinition struct {
	Name        string
	Version     int
	StartStepId string

	Steps    []*DefinitionStep
	StepById map[string]*DefinitionStep
}

type DefinitionStep struct {
	Id   string
	Kind StepKind

	Handler  *Handler
	Verifier *Verifier
	Result   *SagaResult

	Inputs  []StepInputParam
	Outputs []StepOutputParam

	Retry    *RetryPolicy
	Recovery *RecoveryPolicy
	Timeout  time.Duration
	Delay    *time.Duration

	CompensateStepId string

	Transitions map[StepOutcome]string
}

type Handler struct {
	Service string
	Method  string
}

type Verifier struct {
	Type       VerifierType
	Datasource string
	Query      string
	Expect     VerifierExpectModel
}

type VerifierExpectModel struct {
	Equals any
}

type RetryPolicy struct {
	MaxAttempts int
	Backoff     RetryBackoffType
	Delay       time.Duration
}

type RecoveryPolicy struct {
	MaxCycles  int
	OnExceeded string
}
