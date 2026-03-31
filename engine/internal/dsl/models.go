package dsl

import "time"

type Definition struct {
	Name        string
	Version     int
	StartStepId string

	Steps    []*Step
	StepById map[string]*Step
}

type Step struct {
	Id   string
	Kind StepKind

	Handler  *Handler
	Verifier *Verifier
	Result   *SagaResult

	Retry   *RetryPolicy
	Timeout time.Duration

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
