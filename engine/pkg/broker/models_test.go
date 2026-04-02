package broker

import "testing"

func TestErrorInfo_String(t *testing.T) {
	tests := []struct {
		name     string
		input    ErrorInfo
		expected string
	}{
		{
			name: "without details",
			input: ErrorInfo{
				Code:      "PAYMENT_FAILED",
				Message:   "payment was declined",
				Retriable: false,
			},
			expected: "[PAYMENT_FAILED] payment was declined",
		},
		{
			name: "with simple details",
			input: ErrorInfo{
				Code:      "PAYMENT_FAILED",
				Message:   "payment was declined",
				Retriable: false,
				Details: map[string]interface{}{
					"orderId": "123",
					"amount":  999,
				},
			},
			expected: "[PAYMENT_FAILED] payment was declined: amount=999, orderId=123",
		},
		{
			name: "with nested details",
			input: ErrorInfo{
				Code:      "INTERNAL_ERROR",
				Message:   "unexpected error",
				Retriable: true,
				Details: map[string]interface{}{
					"meta": map[string]any{
						"provider": "tinkoff",
					},
				},
			},
			expected: `[INTERNAL_ERROR] unexpected error: meta={"provider":"tinkoff"}`,
		},
		{
			name: "with nil detail value",
			input: ErrorInfo{
				Code:    "BAD_REQUEST",
				Message: "invalid input",
				Details: map[string]interface{}{
					"field": nil,
				},
			},
			expected: "[BAD_REQUEST] invalid input: field=null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.String()
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
