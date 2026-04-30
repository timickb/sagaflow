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

	Input  map[string]string
	Output map[string]string

	Retry   *RetryPolicy
	Timeout time.Duration
	Delay   *time.Duration

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
	Checks     []string
}

type RetryPolicy struct {
	MaxAttempts int
	Backoff     RetryBackoffType
	Delay       time.Duration
}
