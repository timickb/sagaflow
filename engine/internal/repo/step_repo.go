package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/engine/internal/repo/dbstruct"
	"github.com/timickb/sagaflow/lib/db"
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
	mapped, mapErr := step.ToDomain()
	if mapErr != nil {
		return nil, false, mapErr
	}
	return mapped, true, nil
}

// Create - создать шаг экземпляра, ожидающий выполнения
func (r *stepRepo) Create(ctx context.Context, dto *domain.StepCreateDto) (*domain.StepView, error) {
	step := dbstruct.NewSagaStep(dto)
	err := r.db.WithTxSupport(ctx).Create(step).Error
	if err != nil {
		return nil, fmt.Errorf("create saga step: %w", err)
	}
	mapped, mapErr := step.ToDomain()
	if mapErr != nil {
		return nil, mapErr
	}
	return mapped, nil
}

// Update - обновить шаг
func (r *stepRepo) Update(ctx context.Context, dto *domain.StepUpdateDto) error {
	updateMap := dbstruct.NewSagaStepUpdatesMap(dto)
	query := r.db.WithTxSupport(ctx).Model(&dbstruct.DBSagaStep{}).
		Where("saga_id = ? AND step_name = ?", dto.InstanceId, dto.StepName).
		Updates(updateMap)
	if query.Error != nil {
		return fmt.Errorf("update saga step: %w", query.Error)
	}
	if query.RowsAffected == 0 {
		return fmt.Errorf("update saga step: step (%v, %s) not found", dto.InstanceId, dto.StepName)
	}
	return nil
}
