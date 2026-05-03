package verification_source

import (
	"reflect"
	"testing"

	"github.com/timickb/sagaflow/engine/internal/domain"
)

func TestSQLSourceBuildQuery(t *testing.T) {
	source := &SQLSource{}

	tests := []struct {
		name      string
		req       *domain.VerificationRequest
		wantQuery string
		wantArgs  []any
		wantErr   bool
	}{
		{
			name: "build query with input param",
			req: &domain.VerificationRequest{
				Query: `
SELECT count()
FROM fct_orders FINAL
WHERE order_id = ${input.order_id}`,
				InitialContext: initialContext1,
				RuntimeContext: emptyContext,
			},
			wantQuery: `
SELECT count()
FROM fct_orders FINAL
WHERE order_id = ?`,
			wantArgs: []any{"order-1"},
		},
		{
			name: "build query with runtime param",
			req: &domain.VerificationRequest{
				Query: `
SELECT count()
FROM fct_orders FINAL
WHERE amount = ${runtime.payment.amount}`,
				InitialContext: emptyContext,
				RuntimeContext: runtimeContext1,
			},
			wantQuery: `
SELECT count()
FROM fct_orders FINAL
WHERE amount = ?`,
			wantArgs: []any{1000},
		},
		{
			name: "build query with multiple params",
			req: &domain.VerificationRequest{
				Query: `
SELECT count()
FROM fct_orders FINAL
WHERE order_id = ${input.order_id}
  AND status = ${runtime.expected_status}
  AND amount = ${runtime.payment.amount}`,
				InitialContext: initialContext1,
				RuntimeContext: runtimeContext2,
			},
			wantQuery: `
SELECT count()
FROM fct_orders FINAL
WHERE order_id = ?
  AND status = ?
  AND amount = ?`,
			wantArgs: []any{"order-1", "PAID", 1000},
		},
		{
			name: "same param used twice",
			req: &domain.VerificationRequest{
				Query: `
SELECT count()
FROM fct_orders FINAL
WHERE order_id = ${input.order_id}
   OR parent_order_id = ${input.order_id}`,
				InitialContext: initialContext1,
				RuntimeContext: emptyContext,
			},
			wantQuery: `
SELECT count()
FROM fct_orders FINAL
WHERE order_id = ?
   OR parent_order_id = ?`,
			wantArgs: []any{"order-1", "order-1"},
		},
		{
			name: "missing input param returns error",
			req: &domain.VerificationRequest{
				Query: `
SELECT count()
FROM fct_orders FINAL
WHERE order_id = ${input.order_id}`,
				InitialContext: emptyContext,
				RuntimeContext: emptyContext,
			},
			wantErr: true,
		},
		{
			name: "missing runtime param returns error",
			req: &domain.VerificationRequest{
				Query: `
SELECT count()
FROM fct_orders FINAL
WHERE amount = ${runtime.payment.amount}`,
				InitialContext: emptyContext,
				RuntimeContext: emptyContext,
			},
			wantErr: true,
		},
		{
			name: "invalid parameter expression returns error",
			req: &domain.VerificationRequest{
				Query: `
SELECT count()
FROM fct_orders FINAL
WHERE order_id = ${unknown.order_id}`,
				InitialContext: initialContext1,
				RuntimeContext: emptyContext,
			},
			wantErr: true,
		},
		{
			name: "query without params",
			req: &domain.VerificationRequest{
				Query: `
SELECT count()
FROM fct_orders FINAL`,
				InitialContext: emptyContext,
				RuntimeContext: emptyContext,
			},
			wantQuery: `
SELECT count()
FROM fct_orders FINAL`,
			wantArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, gotArgs, err := source.buildQuery(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotQuery != tt.wantQuery {
				t.Fatalf("unexpected query:\ngot:\n%s\nwant:\n%s", gotQuery, tt.wantQuery)
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Fatalf("unexpected args: got %#v, want %#v", gotArgs, tt.wantArgs)
			}
		})
	}
}
