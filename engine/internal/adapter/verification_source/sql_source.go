package verification_source

import (
	"context"
	"database/sql"

	"github.com/timickb/sagaflow/engine/internal/domain"
)

type SQLSource struct {
	db *sql.DB
}

func (s *SQLSource) Verify(ctx context.Context, req *domain.VerificationRequest) (*domain.VerificationResult, error) {
	panic("implement me")
}
