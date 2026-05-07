package worker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/lib/broker"
)

func TestCalculateNextRetry(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("fixed backoff - first attempt", func(t *testing.T) {
		policy := &domain.RetryPolicy{
			Delay:       1000 * time.Millisecond,
			MaxAttempts: 3,
			Backoff:     domain.RetryBackoffTypeFixed,
		}
		stepView := &domain.StepView{
			Attempt:   1,
			UpdatedAt: baseTime,
		}

		nextRetry := calculateNextRetry(policy, stepView)

		require.Equal(t, baseTime.Add(1000*time.Millisecond), nextRetry)
	})

	t.Run("fixed backoff - multiple attempts", func(t *testing.T) {
		policy := &domain.RetryPolicy{
			Delay:       2000 * time.Millisecond,
			MaxAttempts: 3,
			Backoff:     domain.RetryBackoffTypeFixed,
		}
		stepView := &domain.StepView{
			Attempt:   5,
			UpdatedAt: baseTime,
		}

		nextRetry := calculateNextRetry(policy, stepView)

		require.Equal(t, baseTime.Add(2000*time.Millisecond), nextRetry)
	})

	t.Run("exponential backoff - first attempt", func(t *testing.T) {
		policy := &domain.RetryPolicy{
			Delay:       1000 * time.Millisecond,
			MaxAttempts: 3,
			Backoff:     domain.RetryBackoffTypeExponential,
		}
		stepView := &domain.StepView{
			Attempt:   1,
			UpdatedAt: baseTime,
		}

		nextRetry := calculateNextRetry(policy, stepView)

		require.Equal(t, baseTime.Add(2000*time.Millisecond), nextRetry)
	})

	t.Run("exponential backoff - second attempt", func(t *testing.T) {
		policy := &domain.RetryPolicy{
			Delay:       1000 * time.Millisecond,
			MaxAttempts: 3,
			Backoff:     domain.RetryBackoffTypeExponential,
		}
		stepView := &domain.StepView{
			Attempt:   2,
			UpdatedAt: baseTime,
		}

		nextRetry := calculateNextRetry(policy, stepView)

		require.Equal(t, baseTime.Add(4000*time.Millisecond), nextRetry)
	})

	t.Run("exponential backoff - higher attempt", func(t *testing.T) {
		policy := &domain.RetryPolicy{
			Delay:       500 * time.Millisecond,
			MaxAttempts: 5,
			Backoff:     domain.RetryBackoffTypeExponential,
		}
		stepView := &domain.StepView{
			Attempt:   3,
			UpdatedAt: baseTime,
		}

		nextRetry := calculateNextRetry(policy, stepView)

		require.Equal(t, baseTime.Add(3000*time.Millisecond), nextRetry)
	})

	t.Run("exponential backoff - large delay", func(t *testing.T) {
		policy := &domain.RetryPolicy{
			Delay:       10 * time.Second,
			MaxAttempts: 3,
			Backoff:     domain.RetryBackoffTypeExponential,
		}
		stepView := &domain.StepView{
			Attempt:   4,
			UpdatedAt: baseTime,
		}

		nextRetry := calculateNextRetry(policy, stepView)

		require.Equal(t, baseTime.Add(80*time.Second), nextRetry)
	})
}

func TestMergeStepOutputToContext(t *testing.T) {
	t.Run("no outputs defined - returns original context", func(t *testing.T) {
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{"existing": "value"})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Result: map[string]any{"some": "data"},
		}

		result, err := mergeStepOutputToContext(initialCtx, &domain.DefinitionStep{
			Id:      "test_step",
			Kind:    domain.StepKindAction,
			Outputs: []domain.StepOutputParam{},
		}, event)
		require.NoError(t, err)
		require.Equal(t, initialCtx, result)
	})

	t.Run("merge result parameters", func(t *testing.T) {
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{
			"order_id": "123",
		})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Result: map[string]any{
				"status":  "PAID",
				"amount":  float64(100.50),
				"details": map[string]any{"code": "A1"},
			},
		}

		result, err := mergeStepOutputToContext(initialCtx, &domain.DefinitionStep{
			Id:   "test_step",
			Kind: domain.StepKindAction,
			Outputs: []domain.StepOutputParam{
				{SourceNamespace: domain.StepOutputSourceResult, SourceParam: "status", DestinationParam: "order_status"},
				{SourceNamespace: domain.StepOutputSourceResult, SourceParam: "amount", DestinationParam: "order_amount"},
			},
		}, event)
		require.NoError(t, err)

		status, err := result.Find("order_status")
		require.NoError(t, err)
		require.Equal(t, "PAID", status)

		amount, err := result.Find("order_amount")
		require.NoError(t, err)
		require.Equal(t, float64(100.50), amount)

		orderId, err := result.Find("order_id")
		require.NoError(t, err)
		require.Equal(t, "123", orderId)
	})

	t.Run("merge error parameters", func(t *testing.T) {
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Result: map[string]any{"success": true},
			Error: &broker.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: "Validation failed",
				Details: map[string]any{
					"field":   "email",
					"message": "Invalid format",
				},
			},
		}

		result, err := mergeStepOutputToContext(initialCtx, &domain.DefinitionStep{
			Id:   "verify_step",
			Kind: domain.StepKindVerify,
			Outputs: []domain.StepOutputParam{
				{SourceNamespace: domain.StepOutputSourceError, SourceParam: "field", DestinationParam: "error_field"},
				{SourceNamespace: domain.StepOutputSourceError, SourceParam: "message", DestinationParam: "error_message"},
			},
		}, event)
		require.NoError(t, err)

		field, err := result.Find("error_field")
		require.NoError(t, err)
		require.Equal(t, "email", field)

		msg, err := result.Find("error_message")
		require.NoError(t, err)
		require.Equal(t, "Invalid format", msg)
	})

	t.Run("overwrite existing parameter", func(t *testing.T) {
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{
			"order_status": "pending",
			"order_id":     "123",
		})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Result: map[string]any{
				"order_status": "completed",
			},
		}

		result, err := mergeStepOutputToContext(initialCtx, &domain.DefinitionStep{
			Id:   "finalize_step",
			Kind: domain.StepKindAction,
			Outputs: []domain.StepOutputParam{
				{SourceNamespace: domain.StepOutputSourceResult, SourceParam: "order_status", DestinationParam: "order_status"},
			},
		}, event)
		require.NoError(t, err)

		status, err := result.Find("order_status")
		require.NoError(t, err)
		require.Equal(t, "completed", status, "existing parameter should be overwritten")

		orderId, err := result.Find("order_id")
		require.NoError(t, err)
		require.Equal(t, "123", orderId)
	})

	t.Run("missing result parameter - not added to context", func(t *testing.T) {
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Result: map[string]any{
				"existing_key": "value",
			},
		}

		result, err := mergeStepOutputToContext(initialCtx, &domain.DefinitionStep{
			Id:   "test_step",
			Kind: domain.StepKindAction,
			Outputs: []domain.StepOutputParam{
				{SourceNamespace: domain.StepOutputSourceResult, SourceParam: "nonexistent", DestinationParam: "target"},
			},
		}, event)
		require.NoError(t, err)

		_, err = result.Find("target")
		require.Error(t, err)
	})

	t.Run("mixed result and error parameters", func(t *testing.T) {
		initialCtx, err := domain.NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		event := &broker.SagaStepResultEvent{
			Result: map[string]any{"transaction_id": "txn_123"},
			Error: &broker.ErrorInfo{
				Details: map[string]any{"retry_count": float64(2)},
			},
		}

		result, err := mergeStepOutputToContext(initialCtx, &domain.DefinitionStep{
			Id:   "mixed_step",
			Kind: domain.StepKindAction,
			Outputs: []domain.StepOutputParam{
				{SourceNamespace: domain.StepOutputSourceResult, SourceParam: "transaction_id", DestinationParam: "tx_id"},
				{SourceNamespace: domain.StepOutputSourceError, SourceParam: "retry_count", DestinationParam: "retries"},
			},
		}, event)
		require.NoError(t, err)

		txnId, err := result.Find("tx_id")
		require.NoError(t, err)
		require.Equal(t, "txn_123", txnId)

		retries, err := result.Find("retries")
		require.NoError(t, err)
		require.Equal(t, float64(2), retries)
	})

	t.Run("nested context update via multiple steps", func(t *testing.T) {
		var ctx domain.InstanceContext
		ctx, err := domain.NewJsonInstanceContextFromAny(map[string]any{
			"customer_id": "456",
		})
		require.NoError(t, err)

		event1 := &broker.SagaStepResultEvent{
			Result: map[string]any{"billing_id": "bill_789"},
		}
		ctx, err = mergeStepOutputToContext(ctx, &domain.DefinitionStep{
			Id:   "billing_step",
			Kind: domain.StepKindAction,
			Outputs: []domain.StepOutputParam{
				{SourceNamespace: domain.StepOutputSourceResult, SourceParam: "billing_id", DestinationParam: "billing_id"},
			},
		}, event1)
		require.NoError(t, err)

		event2 := &broker.SagaStepResultEvent{
			Result: map[string]any{"billing_id": "bill_999", "payment_status": "paid"},
		}
		ctx, err = mergeStepOutputToContext(ctx, &domain.DefinitionStep{
			Id:   "payment_step",
			Kind: domain.StepKindAction,
			Outputs: []domain.StepOutputParam{
				{SourceNamespace: domain.StepOutputSourceResult, SourceParam: "billing_id", DestinationParam: "billing_id"},
				{SourceNamespace: domain.StepOutputSourceResult, SourceParam: "payment_status", DestinationParam: "payment_status"},
			},
		}, event2)
		require.NoError(t, err)

		billingId, err := ctx.Find("billing_id")
		require.NoError(t, err)
		require.Equal(t, "bill_999", billingId, "billing_id should be overwritten with new value")

		customerId, err := ctx.Find("customer_id")
		require.NoError(t, err)
		require.Equal(t, "456", customerId)

		paymentStatus, err := ctx.Find("payment_status")
		require.NoError(t, err)
		require.Equal(t, "paid", paymentStatus)
	})
}
