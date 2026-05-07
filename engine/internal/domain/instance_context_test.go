package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStepInputContext(t *testing.T) {
	t.Run("empty step inputs", func(t *testing.T) {
		initialCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"order_id": "123",
		})
		require.NoError(t, err)

		runtimeCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"user_id": "456",
		})
		require.NoError(t, err)

		result, err := NewStepInputContext([]StepInputParam{}, initialCtx, runtimeCtx)
		require.NoError(t, err)
		require.NotNil(t, result)

		var got map[string]any
		err = json.Unmarshal(result.GetRaw(), &got)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("single input from initial context", func(t *testing.T) {
		initialCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"customer_name": "John Doe",
			"email":         "john@example.com",
		})
		require.NoError(t, err)

		runtimeCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"session_id": "sess_123",
		})
		require.NoError(t, err)

		stepInputs := []StepInputParam{
			{
				SourceNamespace:  StepInputSourceInputContext,
				SourcePath:       "customer_name",
				DestinationParam: "name",
			},
		}

		result, err := NewStepInputContext(stepInputs, initialCtx, runtimeCtx)
		require.NoError(t, err)
		require.NotNil(t, result)

		name, err := result.Find("name")
		require.NoError(t, err)
		require.Equal(t, "John Doe", name)
	})

	t.Run("single input from runtime context", func(t *testing.T) {
		initialCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"static_value": "original",
		})
		require.NoError(t, err)

		runtimeCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"transaction_id": "txn_789",
			"amount":         float64(99.99),
		})
		require.NoError(t, err)

		stepInputs := []StepInputParam{
			{
				SourceNamespace:  StepInputSourceRuntimeContext,
				SourcePath:       "transaction_id",
				DestinationParam: "txn_id",
			},
		}

		result, err := NewStepInputContext(stepInputs, initialCtx, runtimeCtx)
		require.NoError(t, err)
		require.NotNil(t, result)

		txnId, err := result.Find("txn_id")
		require.NoError(t, err)
		require.Equal(t, "txn_789", txnId)
	})

	t.Run("multiple inputs from both contexts", func(t *testing.T) {
		initialCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"order_id":    "order_123",
			"customer_id": "cust_456",
		})
		require.NoError(t, err)

		runtimeCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"status":    "pending",
			"timestamp": float64(1700000000),
		})
		require.NoError(t, err)

		stepInputs := []StepInputParam{
			{
				SourceNamespace:  StepInputSourceInputContext,
				SourcePath:       "order_id",
				DestinationParam: "id",
			},
			{
				SourceNamespace:  StepInputSourceRuntimeContext,
				SourcePath:       "status",
				DestinationParam: "current_status",
			},
		}

		result, err := NewStepInputContext(stepInputs, initialCtx, runtimeCtx)
		require.NoError(t, err)
		require.NotNil(t, result)

		id, err := result.Find("id")
		require.NoError(t, err)
		require.Equal(t, "order_123", id)

		status, err := result.Find("current_status")
		require.NoError(t, err)
		require.Equal(t, "pending", status)
	})

	t.Run("nested path in initial context", func(t *testing.T) {
		initialCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"user": map[string]any{
				"profile": map[string]any{
					"name": "Alice",
					"age":  float64(30),
				},
			},
		})
		require.NoError(t, err)

		runtimeCtx, err := NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		stepInputs := []StepInputParam{
			{
				SourceNamespace:  StepInputSourceInputContext,
				SourcePath:       "user.profile.name",
				DestinationParam: "user_name",
			},
		}

		result, err := NewStepInputContext(stepInputs, initialCtx, runtimeCtx)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify using Find() method
		userName, err := result.Find("user_name")
		require.NoError(t, err)
		require.Equal(t, "Alice", userName)
	})

	t.Run("nested path in runtime context", func(t *testing.T) {
		initialCtx, err := NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		runtimeCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"data": map[string]any{
				"items": []any{"item1", "item2", "item3"},
			},
		})
		require.NoError(t, err)

		stepInputs := []StepInputParam{
			{
				SourceNamespace:  StepInputSourceRuntimeContext,
				SourcePath:       "data.items.1",
				DestinationParam: "second_item",
			},
		}

		result, err := NewStepInputContext(stepInputs, initialCtx, runtimeCtx)
		require.NoError(t, err)
		require.NotNil(t, result)

		secondItem, err := result.Find("second_item")
		require.NoError(t, err)
		require.Equal(t, "item2", secondItem)
	})

	t.Run("missing path in initial context returns error", func(t *testing.T) {
		initialCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"existing_key": "value",
		})
		require.NoError(t, err)

		runtimeCtx, err := NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		stepInputs := []StepInputParam{
			{
				SourceNamespace:  StepInputSourceInputContext,
				SourcePath:       "nonexistent_key",
				DestinationParam: "target",
			},
		}

		result, err := NewStepInputContext(stepInputs, initialCtx, runtimeCtx)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "find step input value in initial context")
	})

	t.Run("missing path in runtime context returns error", func(t *testing.T) {
		initialCtx, err := NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		runtimeCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"existing_key": "value",
		})
		require.NoError(t, err)

		stepInputs := []StepInputParam{
			{
				SourceNamespace:  StepInputSourceRuntimeContext,
				SourcePath:       "nonexistent_key",
				DestinationParam: "target",
			},
		}

		result, err := NewStepInputContext(stepInputs, initialCtx, runtimeCtx)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "find step input value in initial context")
	})

	t.Run("invalid source namespace returns error", func(t *testing.T) {
		initialCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"key": "value",
		})
		require.NoError(t, err)

		runtimeCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"key": "value",
		})
		require.NoError(t, err)

		stepInputs := []StepInputParam{
			{
				SourceNamespace:  "invalid_source",
				SourcePath:       "key",
				DestinationParam: "target",
			},
		}

		result, err := NewStepInputContext(stepInputs, initialCtx, runtimeCtx)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "invalid step input source")
	})

	t.Run("array index access in runtime context", func(t *testing.T) {
		initialCtx, err := NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		runtimeCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"items": []any{"first", "second", "third"},
		})
		require.NoError(t, err)

		stepInputs := []StepInputParam{
			{
				SourceNamespace:  StepInputSourceRuntimeContext,
				SourcePath:       "items.0",
				DestinationParam: "first_item",
			},
			{
				SourceNamespace:  StepInputSourceRuntimeContext,
				SourcePath:       "items.2",
				DestinationParam: "third_item",
			},
		}

		result, err := NewStepInputContext(stepInputs, initialCtx, runtimeCtx)
		require.NoError(t, err)
		require.NotNil(t, result)

		first, err := result.Find("first_item")
		require.NoError(t, err)
		require.Equal(t, "first", first)

		third, err := result.Find("third_item")
		require.NoError(t, err)
		require.Equal(t, "third", third)
	})

	t.Run("complex nested data structure", func(t *testing.T) {
		initialCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"billing": map[string]any{
				"address": map[string]any{
					"street":  "123 Main St",
					"city":    "Metropolis",
					"zipcode": "12345",
				},
			},
		})
		require.NoError(t, err)

		runtimeCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"order": map[string]any{
				"total":  float64(150.00),
				"status": "processing",
			},
		})
		require.NoError(t, err)

		stepInputs := []StepInputParam{
			{
				SourceNamespace:  StepInputSourceInputContext,
				SourcePath:       "billing.address.street",
				DestinationParam: "delivery_address",
			},
			{
				SourceNamespace:  StepInputSourceRuntimeContext,
				SourcePath:       "order.total",
				DestinationParam: "order_amount",
			},
			{
				SourceNamespace:  StepInputSourceRuntimeContext,
				SourcePath:       "order.status",
				DestinationParam: "status",
			},
		}

		result, err := NewStepInputContext(stepInputs, initialCtx, runtimeCtx)
		require.NoError(t, err)
		require.NotNil(t, result)

		addr, err := result.Find("delivery_address")
		require.NoError(t, err)
		require.Equal(t, "123 Main St", addr)

		amount, err := result.Find("order_amount")
		require.NoError(t, err)
		require.Equal(t, float64(150.00), amount)

		status, err := result.Find("status")
		require.NoError(t, err)
		require.Equal(t, "processing", status)
	})

	t.Run("array index out of range returns error", func(t *testing.T) {
		initialCtx, err := NewJsonInstanceContextFromAny(map[string]any{})
		require.NoError(t, err)

		runtimeCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"items": []any{"only_one"},
		})
		require.NoError(t, err)

		stepInputs := []StepInputParam{
			{
				SourceNamespace:  StepInputSourceRuntimeContext,
				SourcePath:       "items.5",
				DestinationParam: "target",
			},
		}

		result, err := NewStepInputContext(stepInputs, initialCtx, runtimeCtx)
		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("mixed types in result", func(t *testing.T) {
		initialCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"user_id":   "user_123",
			"is_active": true,
			"score":     float64(100),
		})
		require.NoError(t, err)

		runtimeCtx, err := NewJsonInstanceContextFromAny(map[string]any{
			"metadata": map[string]any{
				"source": "api",
			},
		})
		require.NoError(t, err)

		stepInputs := []StepInputParam{
			{
				SourceNamespace:  StepInputSourceInputContext,
				SourcePath:       "user_id",
				DestinationParam: "uid",
			},
			{
				SourceNamespace:  StepInputSourceInputContext,
				SourcePath:       "is_active",
				DestinationParam: "active",
			},
			{
				SourceNamespace:  StepInputSourceInputContext,
				SourcePath:       "score",
				DestinationParam: "points",
			},
			{
				SourceNamespace:  StepInputSourceRuntimeContext,
				SourcePath:       "metadata.source",
				DestinationParam: "origin",
			},
		}

		result, err := NewStepInputContext(stepInputs, initialCtx, runtimeCtx)
		require.NoError(t, err)
		require.NotNil(t, result)

		uid, err := result.Find("uid")
		require.NoError(t, err)
		require.Equal(t, "user_123", uid)

		active, err := result.Find("active")
		require.NoError(t, err)
		require.Equal(t, true, active)

		points, err := result.Find("points")
		require.NoError(t, err)
		require.Equal(t, float64(100), points)

		origin, err := result.Find("origin")
		require.NoError(t, err)
		require.Equal(t, "api", origin)
	})
}
