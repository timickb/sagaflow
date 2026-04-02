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

// Create - создать шаг экземпляра, ожидающий выполнения
func (r *stepRepo) Create(ctx context.Context, dto *domain.StepCreateDto) error {
	err := r.db.WithTxSupport(ctx).Create(dbstruct.NewSagaStep(dto)).Error
	if err != nil {
		return fmt.Errorf("create saga step: %w", err)
	}
	return nil
}

// Update - обновить шаг
func (r *stepRepo) Update(ctx context.Context, dto *domain.StepUpdateDto) error {
	updateMap := dbstruct.NewSagaStepUpdatesMap(dto)
	step := &dbstruct.DBSagaStep{
		SagaId:   dto.InstanceId,
		StepName: dto.StepName,
	}
	query := r.db.WithTxSupport(ctx).Model(step).Updates(updateMap)
	if query.Error != nil {
		return fmt.Errorf("update saga step: %w", query.Error)
	}
	if query.RowsAffected == 0 {
		return fmt.Errorf("update saga step: step (%v, %s) not found", dto.InstanceId, dto.StepName)
	}
	return nil
}
