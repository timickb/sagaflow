package verification_source

import (
	"reflect"
	"testing"

	"github.com/timickb/sagaflow/engine/internal/domain"
)

var (
	emptyContext, _    = domain.NewJsonInstanceContextFromAny(map[string]any{})
	initialContext1, _ = domain.NewJsonInstanceContextFromAny(map[string]any{
		"order_id": "order-1",
	})
	initialContext2, _ = domain.NewJsonInstanceContextFromAny(map[string]any{
		"order": map[string]any{
			"id": "order-1",
		},
	})
	runtimeContext1, _ = domain.NewJsonInstanceContextFromAny(map[string]any{
		"payment": map[string]any{
			"amount": 1000,
		},
	})
	runtimeContext2, _ = domain.NewJsonInstanceContextFromAny(map[string]any{
		"expected_status": "PAID",
		"payment": map[string]any{
			"amount": 1000,
		},
	})
)

func TestFindVerificationParam(t *testing.T) {
	tests := []struct {
		name    string
		req     *domain.VerificationRequest
		source  string
		path    string
		want    any
		wantErr bool
	}{
		{
			name: "find input param",
			req: &domain.VerificationRequest{
				InitialContext: initialContext1,
				RuntimeContext: emptyContext,
			},
			source: "input",
			path:   "order_id",
			want:   "order-1",
		},
		{
			name: "find nested input param",
			req: &domain.VerificationRequest{
				InitialContext: initialContext2,
				RuntimeContext: emptyContext,
			},
			source: "input",
			path:   "order.id",
			want:   "order-1",
		},
		{
			name: "find runtime param",
			req: &domain.VerificationRequest{
				InitialContext: emptyContext,
				RuntimeContext: runtimeContext1,
			},
			source: "runtime",
			path:   "payment.amount",
			want:   1000,
		},
		{
			name: "input param not found",
			req: &domain.VerificationRequest{
				InitialContext: emptyContext,
				RuntimeContext: emptyContext,
			},
			source:  "input",
			path:    "order_id",
			wantErr: true,
		},
		{
			name: "runtime param not found",
			req: &domain.VerificationRequest{
				InitialContext: emptyContext,
				RuntimeContext: emptyContext,
			},
			source:  "runtime",
			path:    "payment.amount",
			wantErr: true,
		},
		{
			name: "unsupported source",
			req: &domain.VerificationRequest{
				InitialContext: emptyContext,
				RuntimeContext: emptyContext,
			},
			source:  "unknown",
			path:    "order_id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findVerificationParam(tt.req, tt.source, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected value: got %v, want %v", got, tt.want)
			}
		})
	}
}
