package dsl

import "github.com/timickb/sagaflow/engine/internal/domain"

type RawSagaDefinition struct {
	Saga  RawSagaMeta `yaml:"saga"`
	Steps []RawStep   `yaml:"steps"`
}

type RawSagaMeta struct {
	Name    string `yaml:"name"`
	Version int    `yaml:"version"`
	Start   string `yaml:"start"`
}

type RawStep struct {
	Id         string             `yaml:"id"`
	Kind       domain.StepKind    `yaml:"kind"`
	Handler    *RawHandler        `yaml:"handler,omitempty"`
	Verifier   *RawVerifier       `yaml:"verifier,omitempty"`
	Retry      *RawRetryPolicy    `yaml:"retry,omitempty"`
	Timeout    string             `yaml:"timeout,omitempty"`
	Compensate *string            `yaml:"compensate,omitempty"`
	Result     *domain.SagaResult `yaml:"result,omitempty"`
	On         map[string]string  `yaml:"on,omitempty"`
}

type RawHandler struct {
	Service string `yaml:"service"`
	Method  string `yaml:"method"`
}

type RawVerifier struct {
	Type       domain.VerifierType `yaml:"type"`
	Datasource string              `yaml:"datasource,omitempty"`
	Query      string              `yaml:"query,omitempty"`
	Checks     []string            `yaml:"checks,omitempty"`
}

type RawRetryPolicy struct {
	MaxAttempts int                     `yaml:"max_attempts"`
	Backoff     domain.RetryBackoffType `yaml:"backoff,omitempty"`
	Delay       string                  `yaml:"delay,omitempty"`
}
