package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/engine/internal/repo/dbstruct"
	"github.com/timickb/sagaflow/engine/pkg/db"
	"gorm.io/gorm"
)

type stepRepo struct {
	db *db.Database
}

func NewStepRepo(db *db.Database) *stepRepo {
	return &stepRepo{
		db: db,
	}
}

func (r *stepRepo) GetByInstanceAndName(
	ctx context.Context, instanceId uuid.UUID, stepName string,
) (*domain.StepView, bool, error) {
	var step dbstruct.DBSagaStep

	err := r.db.WithTxSupport(ctx).Where("saga_id = ? AND step_name = ?", instanceId, stepName).
		Take(&step).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get step by instance and name: %w", err)
	}
	return step.ToDomain(), true, nil
}
