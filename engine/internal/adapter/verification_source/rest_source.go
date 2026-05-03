package verification_source

import (
	"context"
	"net/http"

	"github.com/timickb/sagaflow/engine/internal/domain"
)

type RESTSource struct {
	client  *http.Client
	baseUrl string
	headers map[string]string
}

func (s *RESTSource) Verify(ctx context.Context, req *domain.VerificationRequest) (*domain.VerificationResult, error) {
	// TODO: implement
	panic("implement me")
}
