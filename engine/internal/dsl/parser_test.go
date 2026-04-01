package dsl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/engine/pkg/utils"
)

func TestRawSagaDefinition_ValidateAndNormalize(t *testing.T) {
	t.Run("valid saga definition", func(t *testing.T) {
		raw := &RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "test-saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "10s",
					Handler: &RawHandler{Service: "svc1", Method: "Do"},
					On:      map[string]string{"committed": "step2"},
				},
				{
					Id:      "step2",
					Kind:    domain.StepKindTerminal,
					Timeout: "5s",
					Result:  utils.Ptr(domain.SagaResultCompleted),
					On:      map[string]string{},
				},
			},
		}

		def, err := raw.ValidateAndNormalize()
		require.NoError(t, err)
		assert.Equal(t, "test-saga", def.Name)
		assert.Equal(t, 1, def.Version)
		assert.Equal(t, "step1", def.StartStepId)
		assert.Len(t, def.Steps, 2)
		assert.Contains(t, def.StepById, "step1")
		assert.Contains(t, def.StepById, "step2")
	})

	t.Run("invalid version - zero", func(t *testing.T) {
		raw := &RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "test-saga",
				Version: 0,
				Start:   "step1",
			},
			Steps: []RawStep{
				{Id: "step1", Kind: domain.StepKindAction, Timeout: "10s", Handler: &RawHandler{Service: "svc", Method: "Do"}},
			},
		}

		_, err := raw.ValidateAndNormalize()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version must be positive")
	})

	t.Run("invalid version - negative", func(t *testing.T) {
		raw := &RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "test-saga",
				Version: -1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{Id: "step1", Kind: domain.StepKindAction, Timeout: "10s", Handler: &RawHandler{Service: "svc", Method: "Do"}},
			},
		}

		_, err := raw.ValidateAndNormalize()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version must be positive")
	})

	t.Run("empty name", func(t *testing.T) {
		raw := &RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{Id: "step1", Kind: domain.StepKindAction, Timeout: "10s", Handler: &RawHandler{Service: "svc", Method: "Do"}},
			},
		}

		_, err := raw.ValidateAndNormalize()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name could not be empty")
	})

	t.Run("start step not declared", func(t *testing.T) {
		raw := &RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "test-saga",
				Version: 1,
				Start:   "nonexistent",
			},
			Steps: []RawStep{
				{Id: "step1", Kind: domain.StepKindAction, Timeout: "10s", Handler: &RawHandler{Service: "svc", Method: "Do"}},
			},
		}

		_, err := raw.ValidateAndNormalize()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start step nonexistent does not declared")
	})

	t.Run("next step not declared", func(t *testing.T) {
		raw := &RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "test-saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "10s",
					Handler: &RawHandler{Service: "svc", Method: "Do"},
					On:      map[string]string{"committed": "nonexistent"},
				},
			},
		}

		_, err := raw.ValidateAndNormalize()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "next step step1 does not declared")
	})

	t.Run("parse step error", func(t *testing.T) {
		raw := &RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "test-saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "invalid",
					Handler: &RawHandler{Service: "svc", Method: "Do"},
					On:      map[string]string{},
				},
			},
		}

		_, err := raw.ValidateAndNormalize()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse step")
	})
}

func TestParseStep(t *testing.T) {
	t.Run("valid action step", func(t *testing.T) {
		raw := RawStep{
			Id:      "action1",
			Kind:    domain.StepKindAction,
			Timeout: "10s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			On:      map[string]string{"committed": "next1"},
		}

		step, err := parseStep(raw)
		require.NoError(t, err)
		assert.Equal(t, "action1", step.Id)
		assert.Equal(t, domain.StepKindAction, step.Kind)
		assert.Equal(t, 10*time.Second, step.Timeout)
		assert.NotNil(t, step.Handler)
		assert.Equal(t, "svc1", step.Handler.Service)
		assert.Equal(t, "Do", step.Handler.Method)
		assert.Equal(t, "next1", step.Transitions[domain.OutcomeCommitted])
	})

	t.Run("valid verify step", func(t *testing.T) {
		raw := RawStep{
			Id:      "verify1",
			Kind:    domain.StepKindVerify,
			Timeout: "5s",
			Verifier: &RawVerifier{
				Type:       domain.VerifierTypeSql,
				Datasource: "db1",
				Query:      "SELECT COUNT(*) FROM orders",
				Checks:     []string{"count > 0"},
			},
			On: map[string]string{"passed": "next1", "verification_failed": "step2"},
		}

		step, err := parseStep(raw)
		require.NoError(t, err)
		assert.Equal(t, "verify1", step.Id)
		assert.Equal(t, domain.StepKindVerify, step.Kind)
		assert.NotNil(t, step.Verifier)
		assert.Equal(t, domain.VerifierTypeSql, step.Verifier.Type)
	})

	t.Run("valid reconcile step", func(t *testing.T) {
		raw := RawStep{
			Id:      "reconcile1",
			Kind:    domain.StepKindReconcile,
			Timeout: "30s",
			Handler: &RawHandler{Service: "svc1", Method: "Reconcile"},
			On:      map[string]string{"committed": "next1"},
		}

		step, err := parseStep(raw)
		require.NoError(t, err)
		assert.Equal(t, domain.StepKindReconcile, step.Kind)
	})

	t.Run("valid terminal step with COMPENSATED result", func(t *testing.T) {
		raw := RawStep{
			Id:      "terminal1",
			Kind:    domain.StepKindTerminal,
			Timeout: "1s",
			Result:  utils.Ptr(domain.SagaResultCompensated),
			On:      map[string]string{},
		}

		step, err := parseStep(raw)
		require.NoError(t, err)
		assert.Equal(t, domain.StepKindTerminal, step.Kind)
		assert.NotNil(t, step.Result)
		assert.Equal(t, domain.SagaResultCompensated, *step.Result)
	})

	t.Run("valid step with retry policy", func(t *testing.T) {
		raw := RawStep{
			Id:      "retry1",
			Kind:    domain.StepKindAction,
			Timeout: "10s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			Retry: &RawRetryPolicy{
				MaxAttempts: 3,
				Backoff:     domain.RetryBackoffTypeExponential,
				Delay:       "2s",
			},
			On: map[string]string{"failed": "retry1", "committed": "next1"},
		}

		step, err := parseStep(raw)
		require.NoError(t, err)
		assert.NotNil(t, step.Retry)
		assert.Equal(t, 3, step.Retry.MaxAttempts)
		assert.Equal(t, domain.RetryBackoffTypeExponential, step.Retry.Backoff)
		assert.Equal(t, 2*time.Second, step.Retry.Delay)
	})

	t.Run("empty step id", func(t *testing.T) {
		raw := RawStep{
			Id:      "",
			Kind:    domain.StepKindAction,
			Timeout: "10s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			On:      map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id could not be empty")
	})

	t.Run("invalid timeout format", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindAction,
			Timeout: "invalid",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			On:      map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse step timeout")
	})

	t.Run("zero timeout", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindAction,
			Timeout: "0s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			On:      map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout is zero")
	})

	t.Run("action step without handler", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindAction,
			Timeout: "10s",
			On:      map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler is required")
	})

	t.Run("action step with unexpected verifier", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindAction,
			Timeout: "10s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			Verifier: &RawVerifier{
				Type: domain.VerifierTypeSql,
			},
			On: map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected verifier")
	})

	t.Run("action step with unexpected result", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindAction,
			Timeout: "10s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			Result:  utils.Ptr(domain.SagaResultCompleted),
			On:      map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected result")
	})

	t.Run("verify step without verifier", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindVerify,
			Timeout: "10s",
			On:      map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "verifier is required")
	})

	t.Run("verify step with unexpected handler", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindVerify,
			Timeout: "10s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			Verifier: &RawVerifier{
				Type: domain.VerifierTypeSql,
			},
			On: map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected handler")
	})

	t.Run("verify step with unexpected result", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindVerify,
			Timeout: "10s",
			Verifier: &RawVerifier{
				Type: domain.VerifierTypeSql,
			},
			Result: utils.Ptr(domain.SagaResultCompleted),
			On:     map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected result")
	})

	t.Run("verify step with invalid verifier type", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindVerify,
			Timeout: "10s",
			Verifier: &RawVerifier{
				Type: "invalid",
			},
			On: map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid verifier type")
	})

	t.Run("terminal step without result", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindTerminal,
			Timeout: "1s",
			On:      map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid result value")
	})

	t.Run("terminal step with invalid result", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindTerminal,
			Timeout: "1s",
			Result:  utils.Ptr(domain.SagaResult("INVALID")),
			On:      map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid result value")
	})

	t.Run("terminal step with unexpected handler", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindTerminal,
			Timeout: "1s",
			Result:  utils.Ptr(domain.SagaResultCompleted),
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			On:      map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected handler")
	})

	t.Run("terminal step with unexpected verifier", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindTerminal,
			Timeout: "1s",
			Result:  utils.Ptr(domain.SagaResultCompleted),
			Verifier: &RawVerifier{
				Type: domain.VerifierTypeSql,
			},
			On: map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected verifier")
	})

	t.Run("invalid step outcome", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindAction,
			Timeout: "10s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			On:      map[string]string{"invalid_outcome": "next1"},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid step outcome")
	})

	t.Run("empty next step", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindAction,
			Timeout: "10s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			On:      map[string]string{"committed": ""},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty next step")
	})

	t.Run("retry with invalid backoff", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindAction,
			Timeout: "10s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			Retry: &RawRetryPolicy{
				MaxAttempts: 3,
				Backoff:     "invalid",
				Delay:       "1s",
			},
			On: map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid step retry backoff")
	})

	t.Run("retry with zero max attempts", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindAction,
			Timeout: "10s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			Retry: &RawRetryPolicy{
				MaxAttempts: 0,
				Backoff:     domain.RetryBackoffTypeFixed,
				Delay:       "1s",
			},
			On: map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "retry max attempts must be positive")
	})

	t.Run("retry with negative max attempts", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindAction,
			Timeout: "10s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			Retry: &RawRetryPolicy{
				MaxAttempts: -1,
				Backoff:     domain.RetryBackoffTypeFixed,
				Delay:       "1s",
			},
			On: map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "retry max attempts must be positive")
	})

	t.Run("retry with invalid delay", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindAction,
			Timeout: "10s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			Retry: &RawRetryPolicy{
				MaxAttempts: 3,
				Backoff:     domain.RetryBackoffTypeFixed,
				Delay:       "invalid",
			},
			On: map[string]string{},
		}

		_, err := parseStep(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse step retry delay")
	})

	t.Run("all valid step outcomes", func(t *testing.T) {
		raw := RawStep{
			Id:      "step1",
			Kind:    domain.StepKindAction,
			Timeout: "10s",
			Handler: &RawHandler{Service: "svc1", Method: "Do"},
			On: map[string]string{
				"committed":           "step2",
				"rejected":            "step3",
				"failed":              "step4",
				"timeout":             "step5",
				"passed":              "step6",
				"verification_failed": "step7",
			},
		}

		step, err := parseStep(raw)
		require.NoError(t, err)
		assert.Len(t, step.Transitions, 6)
	})
}

func TestEnums(t *testing.T) {
	t.Run("StepOutcome IsValid", func(t *testing.T) {
		assert.True(t, domain.OutcomeCommitted.IsValid())
		assert.True(t, domain.OutcomeRejected.IsValid())
		assert.True(t, domain.OutcomeFailed.IsValid())
		assert.True(t, domain.OutcomeTimeout.IsValid())
		assert.True(t, domain.OutcomePassed.IsValid())
		assert.True(t, domain.OutcomeVerificationFailed.IsValid())
		assert.False(t, domain.StepOutcome("invalid").IsValid())
	})

	t.Run("SagaResult IsValid", func(t *testing.T) {
		assert.True(t, domain.SagaResultCompleted.IsValid())
		assert.True(t, domain.SagaResultFailed.IsValid())
		assert.True(t, domain.SagaResultCompensated.IsValid())
		assert.True(t, domain.SagaResultInconsistent.IsValid())
		assert.False(t, domain.SagaResult("invalid").IsValid())
	})

	t.Run("VerifierType IsValid", func(t *testing.T) {
		assert.True(t, domain.VerifierTypeSql.IsValid())
		assert.True(t, domain.VerifierTypeComposite.IsValid())
		assert.True(t, domain.VerifierTypeApi.IsValid())
		assert.False(t, domain.VerifierType("invalid").IsValid())
	})

	t.Run("RetryBackoffType IsValid", func(t *testing.T) {
		assert.True(t, domain.RetryBackoffTypeFixed.IsValid())
		assert.True(t, domain.RetryBackoffTypeExponential.IsValid())
		assert.False(t, domain.RetryBackoffType("invalid").IsValid())
	})
}
