package verification_source

import (
	"context"

	"github.com/timickb/sagaflow/engine/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoSource struct {
	client *mongo.Client
}

func (s *MongoSource) Verify(ctx context.Context, req *domain.VerificationRequest) (*domain.VerificationResult, error) {
	panic("implement me")
}
