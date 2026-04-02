package domain

import (
	"encoding/json"
	"testing"
)

func TestJsonInstanceContext_AppendMap(t *testing.T) {
	tests := []struct {
		name        string
		raw         json.RawMessage
		appendData  map[string]any
		expected    map[string]any
		expectError bool
	}{
		{
			name:       "append to existing json",
			raw:        json.RawMessage(`{"foo":"bar","count":1}`),
			appendData: map[string]any{"baz": true},
			expected: map[string]any{
				"foo":   "bar",
				"count": float64(1),
				"baz":   true,
			},
		},
		{
			name:       "override existing key",
			raw:        json.RawMessage(`{"foo":"bar","count":1}`),
			appendData: map[string]any{"count": 2},
			expected: map[string]any{
				"foo":   "bar",
				"count": float64(2),
			},
		},
		{
			name:       "empty raw",
			raw:        nil,
			appendData: map[string]any{"foo": "bar"},
			expected: map[string]any{
				"foo": "bar",
			},
		},
		{
			name:       "null raw",
			raw:        json.RawMessage(`null`),
			appendData: map[string]any{"foo": "bar"},
			expected: map[string]any{
				"foo": "bar",
			},
		},
		{
			name:        "invalid raw json",
			raw:         json.RawMessage(`{invalid json}`),
			appendData:  map[string]any{"foo": "bar"},
			expectError: true,
		},
		{
			name:       "empty append map",
			raw:        json.RawMessage(`{"foo":"bar"}`),
			appendData: map[string]any{},
			expected: map[string]any{
				"foo": "bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewJsonInstanceContextFromRaw(tt.raw)

			result, err := ctx.AppendMap(tt.appendData)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}

			var got map[string]any
			if err := json.Unmarshal(result.GetRaw(), &got); err != nil {
				t.Fatalf("failed to unmarshal result raw: %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Fatalf("expected len=%d, got len=%d; got=%v", len(tt.expected), len(got), got)
			}

			for k, expectedValue := range tt.expected {
				gotValue, ok := got[k]
				if !ok {
					t.Fatalf("expected key %q to exist", k)
				}
				if !valuesEqual(gotValue, expectedValue) {
					t.Fatalf("expected key %q to be %v, got %v", k, expectedValue, gotValue)
				}
			}
		})
	}
}

func TestJsonInstanceContext_AppendMap_ReturnsNewContext(t *testing.T) {
	ctx := NewJsonInstanceContextFromRaw(json.RawMessage(`{"foo":"bar"}`))

	result, err := ctx.AppendMap(map[string]any{"baz": 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(ctx.GetRaw()) != `{"foo":"bar"}` {
		t.Fatalf("original context must not change, got %s", string(ctx.GetRaw()))
	}

	var got map[string]any
	if err := json.Unmarshal(result.GetRaw(), &got); err != nil {
		t.Fatalf("failed to unmarshal result raw: %v", err)
	}

	if got["foo"] != "bar" {
		t.Fatalf("expected foo=bar, got %v", got["foo"])
	}
	if got["baz"] != float64(1) {
		t.Fatalf("expected baz=1, got %v", got["baz"])
	}
}

func valuesEqual(a, b any) bool {
	switch av := a.(type) {
	case float64:
		if bv, ok := b.(float64); ok {
			return av == bv
		}
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
	case nil:
		return b == nil
	}
	return a == b
}
