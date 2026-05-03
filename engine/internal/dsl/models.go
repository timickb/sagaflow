package dsl

import "github.com/timickb/sagaflow/engine/internal/domain"

type RawSagaDefinition struct {
	Saga   RawSagaMeta             `yaml:"saga"`
	Inputs map[string]RawInputSpec `yaml:"inputs"`
	Steps  []RawStep               `yaml:"steps"`
}

type RawSagaMeta struct {
	Name           string `yaml:"name"`
	Version        int    `yaml:"version"`
	Start          string `yaml:"start"`
	IdempotencyKey string `yaml:"idempotency_key"`
}

type RawInputSpec struct {
	Type        string                  `yaml:"type"`
	Required    bool                    `yaml:"required,omitempty"`
	Description string                  `yaml:"description,omitempty"`
	Default     interface{}             `yaml:"default,omitempty"`
	Nullable    bool                    `yaml:"nullable,omitempty"`
	Enum        []interface{}           `yaml:"enum,omitempty"`
	Properties  map[string]RawInputSpec `yaml:"properties,omitempty"`
	Items       *RawInputSpec           `yaml:"items,omitempty"`
}

type RawStep struct {
	Id         string             `yaml:"id"`
	Kind       domain.StepKind    `yaml:"kind"`
	Handler    *RawHandler        `yaml:"handler,omitempty"`
	Verifier   *RawVerifier       `yaml:"verifier,omitempty"`
	Retry      *RawRetryPolicy    `yaml:"retry,omitempty"`
	Input      map[string]string  `yaml:"input,omitempty"`
	Output     map[string]string  `yaml:"output,omitempty"`
	Timeout    string             `yaml:"timeout,omitempty"`
	Delay      *string            `yaml:"delay,omitempty"`
	Compensate *string            `yaml:"compensate,omitempty"`
	Result     *domain.SagaResult `yaml:"result,omitempty"`
	On         map[string]string  `yaml:"on,omitempty"`
}

type RawHandler struct {
	Service string `yaml:"service"`
	Method  string `yaml:"method"`
}

type RawVerifier struct {
	Type       domain.VerifierType    `yaml:"type"`
	Datasource string                 `yaml:"datasource,omitempty"`
	Query      string                 `yaml:"query,omitempty"`
	Checks     []string               `yaml:"checks,omitempty"`
	Expect     RawVerifierExpectModel `yaml:"expect,omitempty"`
}

type RawVerifierExpectModel struct {
	Equals any `json:"equals,omitempty"`
}

type RawRetryPolicy struct {
	MaxAttempts int                     `yaml:"max_attempts"`
	Backoff     domain.RetryBackoffType `yaml:"backoff,omitempty"`
	Delay       string                  `yaml:"delay,omitempty"`
}
