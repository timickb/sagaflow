package dsl

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/utils"
)

func TestValidateAndNormalize(t *testing.T) {
	t.Run("invalid version zero", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 0,
				Start:   "step1",
			},
			Steps: []RawStep{
				{Id: "step1", Kind: domain.StepKindTerminal, Timeout: "1s"},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "definitions version must be positive number")
	})

	t.Run("invalid version negative", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: -1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{Id: "step1", Kind: domain.StepKindTerminal, Timeout: "1s"},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "definitions version must be positive number")
	})

	t.Run("empty saga name", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{Id: "step1", Kind: domain.StepKindTerminal, Timeout: "1s"},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "saga name could not be empty")
	})

	t.Run("start step not declared", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "nonexistent",
			},
			Steps: []RawStep{
				{Id: "step1", Kind: domain.StepKindTerminal, Timeout: "1s"},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "start step nonexistent does not declared")
	})

	t.Run("next step not declared", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "1s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					On:      map[string]string{"committed": "nonexistent"},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "next step step1 does not declared")
	})

	t.Run("valid saga with terminal step", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:     "step1",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
					On:     map[string]string{},
				},
			},
		}
		def, err := raw.ValidateAndNormalize()
		require.NoError(t, err)
		require.NotNil(t, def)
		require.Equal(t, "Test Saga", def.Name)
		require.Equal(t, 1, def.Version)
		require.Equal(t, "step1", def.StartStepId)
		require.Len(t, def.Steps, 1)
	})

	t.Run("valid saga with action and terminal steps", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Order Saga",
				Version: 2,
				Start:   "create_order",
			},
			Steps: []RawStep{
				{
					Id:      "create_order",
					Kind:    domain.StepKindAction,
					Timeout: "30s",
					Handler: &RawHandler{Service: "order", Method: "create"},
					On:      map[string]string{string(domain.OutcomeCommitted): "verify_order"},
				},
				{
					Id:      "verify_order",
					Kind:    domain.StepKindVerify,
					Timeout: "10s",
					Verifier: &RawVerifier{
						Type:  domain.VerifierTypeSql,
						Query: "SELECT ...",
					},
					On: map[string]string{
						string(domain.OutcomeMatched):   "finalize",
						string(domain.OutcomeUnmatched): "fail_saga",
					},
				},
				{
					Id:     "finalize",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
					On:     map[string]string{},
				},
				{
					Id:     "fail_saga",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultFailed),
					On:     map[string]string{},
				},
			},
		}
		def, err := raw.ValidateAndNormalize()
		require.NoError(t, err)
		require.NotNil(t, def)
		require.Equal(t, "Order Saga", def.Name)
		require.Equal(t, 2, def.Version)
		require.Len(t, def.Steps, 4)
		require.NotNil(t, def.StepById)
	})

	t.Run("valid saga with delay", func(t *testing.T) {
		delay := "5s"
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Delayed Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "30s",
					Delay:   &delay,
					Handler: &RawHandler{Service: "svc", Method: "method"},
					On:      map[string]string{string(domain.OutcomeCommitted): "step2"},
				},
				{
					Id:     "step2",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
					On:     map[string]string{},
				},
			},
		}
		def, err := raw.ValidateAndNormalize()
		require.NoError(t, err)
		require.NotNil(t, def)
		require.NotNil(t, def.Steps[0].Delay)
		require.Equal(t, int64(5000000000), def.Steps[0].Delay.Nanoseconds())
	})

	t.Run("valid saga with retry policy", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Retry Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "30s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					Retry: &RawRetryPolicy{
						MaxAttempts: 3,
						Backoff:     domain.RetryBackoffTypeExponential,
						Delay:       "1s",
					},
					On: map[string]string{string(domain.OutcomeCommitted): "step2"},
				},
				{
					Id:     "step2",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
					On:     map[string]string{},
				},
			},
		}
		def, err := raw.ValidateAndNormalize()
		require.NoError(t, err)
		require.NotNil(t, def)
		require.NotNil(t, def.Steps[0].Retry)
		require.Equal(t, 3, def.Steps[0].Retry.MaxAttempts)
		require.Equal(t, domain.RetryBackoffTypeExponential, def.Steps[0].Retry.Backoff)
	})

	t.Run("valid saga with input and output params", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "IO Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "30s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					Input: map[string]string{
						"order_id": "${runtime.orderId}",
					},
					Output: map[string]string{
						"status": "${result.status}",
					},
					On: map[string]string{string(domain.OutcomeCommitted): "step2"},
				},
				{
					Id:     "step2",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
					On:     map[string]string{},
				},
			},
		}
		def, err := raw.ValidateAndNormalize()
		require.NoError(t, err)
		require.NotNil(t, def)
		require.Len(t, def.Steps[0].Inputs, 1)
		require.Len(t, def.Steps[0].Outputs, 1)
		require.Equal(t, domain.StepInputSource("runtime"), def.Steps[0].Inputs[0].SourceNamespace)
		require.Equal(t, "orderId", def.Steps[0].Inputs[0].SourcePath)
		require.Equal(t, domain.StepOutputSource("result"), def.Steps[0].Outputs[0].SourceNamespace)
		require.Equal(t, "status", def.Steps[0].Outputs[0].SourceParam)
	})

	t.Run("valid saga with compensate step", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Compensate Saga",
				Version: 1,
				Start:   "action_step",
			},
			Steps: []RawStep{
				{
					Id:         "action_step",
					Kind:       domain.StepKindAction,
					Timeout:    "30s",
					Handler:    &RawHandler{Service: "svc", Method: "do"},
					Compensate: strPtr("compensate_step"),
					On:         map[string]string{string(domain.OutcomeFailed): "compensate_step"},
				},
				{
					Id:      "compensate_step",
					Kind:    domain.StepKindCompensate,
					Timeout: "10s",
					Handler: &RawHandler{Service: "svc", Method: "undo"},
					On:      map[string]string{string(domain.OutcomeCommitted): "terminal"},
				},
				{
					Id:     "terminal",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompensated),
					On:     map[string]string{},
				},
			},
		}
		def, err := raw.ValidateAndNormalize()
		require.NoError(t, err)
		require.NotNil(t, def)
		require.Equal(t, "compensate_step", def.Steps[0].CompensateStepId)
	})

	t.Run("valid reconcile step with recovery policy", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Reconcile Saga",
				Version: 1,
				Start:   "reconcile_step",
			},
			Steps: []RawStep{
				{
					Id:      "reconcile_step",
					Kind:    domain.StepKindReconcile,
					Timeout: "60s",
					Handler: &RawHandler{Service: "svc", Method: "check"},
					Recovery: &RawRecoveryPolicy{
						MaxCycles:  5,
						OnExceeded: "terminal",
					},
					On: map[string]string{
						string(domain.OutcomeMatched):   "terminal",
						string(domain.OutcomeUnmatched): "terminal",
					},
				},
				{
					Id:     "terminal",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultInconsistent),
					On:     map[string]string{},
				},
			},
		}
		def, err := raw.ValidateAndNormalize()
		require.NoError(t, err)
		require.NotNil(t, def)
		require.NotNil(t, def.Steps[0].Recovery)
		require.Equal(t, 5, def.Steps[0].Recovery.MaxCycles)
		require.Equal(t, "terminal", def.Steps[0].Recovery.OnExceeded)
	})
}

func TestValidateAndNormalize_StepValidation(t *testing.T) {
	t.Run("empty step id", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "",
			},
			Steps: []RawStep{
				{Id: "", Kind: domain.StepKindAction, Timeout: "1s"},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "step id could not be empty")
	})

	t.Run("invalid timeout format", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{Id: "step1", Kind: domain.StepKindAction, Timeout: "invalid"},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "parse step timeout")
	})

	t.Run("zero timeout", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{Id: "step1", Kind: domain.StepKindAction, Timeout: "0s"},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "step timeout is zero")
	})

	t.Run("zero delay", func(t *testing.T) {
		delay := "invalid"
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "1s",
					Delay:   &delay,
					Handler: &RawHandler{Service: "svc", Method: "method"},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "parse step delay")
	})

	t.Run("invalid retry backoff", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "1s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					Retry: &RawRetryPolicy{
						MaxAttempts: 1,
						Backoff:     "invalid",
						Delay:       "1s",
					},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "invalid step retry backoff type")
	})

	t.Run("non-positive retry max attempts", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "1s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					Retry: &RawRetryPolicy{
						MaxAttempts: 0,
						Backoff:     domain.RetryBackoffTypeFixed,
						Delay:       "1s",
					},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "step retry max attempts must be positive")
	})

	t.Run("invalid input source format", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "1s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					Input: map[string]string{
						"param": "invalid",
					},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "invalid input source invalid")
	})

	t.Run("invalid input destination format", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "1s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					Input: map[string]string{
						"123": "${runtime.orderId}",
					},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "invalid input destination 123")
	})

	t.Run("invalid output source format", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "1s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					Output: map[string]string{
						"result": "invalid",
					},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "invalid output source invalid")
	})

	t.Run("invalid output destination format", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "1s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					Output: map[string]string{
						"123": "${result.status}",
					},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "invalid output destination 123")
	})

	t.Run("action step without handler", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{Id: "step1", Kind: domain.StepKindAction, Timeout: "1s"},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "handler is required for action")
	})

	t.Run("action step with unexpected verifier", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "1s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					Verifier: &RawVerifier{
						Type: domain.VerifierTypeSql,
					},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "unexpected verifier for action")
	})

	t.Run("action step with unexpected result", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "1s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					Result:  utils.Ptr(domain.SagaResultCompleted),
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "unexpected result for action")
	})

	t.Run("verify step without verifier", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindVerify,
					Timeout: "10s",
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "verifier is required for verify step")
	})

	t.Run("verify step with unexpected handler", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindVerify,
					Timeout: "10s",
					Verifier: &RawVerifier{
						Type:  domain.VerifierTypeSql,
						Query: "SELECT ...",
					},
					Handler: &RawHandler{Service: "svc", Method: "method"},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "unexpected handler for verify step")
	})

	t.Run("verify step with invalid verifier type", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindVerify,
					Timeout: "10s",
					Verifier: &RawVerifier{
						Type:  "invalid",
						Query: "SELECT ...",
					},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "invalid verifier type")
	})

	t.Run("terminal step without result", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:   "step1",
					Kind: domain.StepKindTerminal,
					On:   map[string]string{},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "invalid result value for terminal step")
	})

	t.Run("terminal step with handler", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindTerminal,
					Result:  utils.Ptr(domain.SagaResultCompleted),
					Handler: &RawHandler{Service: "svc", Method: "method"},
					On:      map[string]string{},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "unexpected handler for terminal step")
	})

	t.Run("reconcile step with recovery zero cycles", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindReconcile,
					Timeout: "10s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					Recovery: &RawRecoveryPolicy{
						MaxCycles:  0,
						OnExceeded: "fail",
					},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "recovery policy requires at least one cycle")
	})

	t.Run("invalid step outcome", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "1s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					On:      map[string]string{"invalid_outcome": "step2"},
				},
				{
					Id:     "step2",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
					On:     map[string]string{},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "invalid step outcome")
	})

	t.Run("empty next step", func(t *testing.T) {
		raw := RawSagaDefinition{
			Saga: RawSagaMeta{
				Name:    "Test Saga",
				Version: 1,
				Start:   "step1",
			},
			Steps: []RawStep{
				{
					Id:      "step1",
					Kind:    domain.StepKindAction,
					Timeout: "1s",
					Handler: &RawHandler{Service: "svc", Method: "method"},
					On:      map[string]string{string(domain.OutcomeCommitted): ""},
				},
				{
					Id:     "step2",
					Kind:   domain.StepKindTerminal,
					Result: utils.Ptr(domain.SagaResultCompleted),
					On:     map[string]string{},
				},
			},
		}
		_, err := raw.ValidateAndNormalize()
		require.ErrorContains(t, err, "unexpected empty next step")
	})
}

func strPtr(s string) *string {
	return &s
}
