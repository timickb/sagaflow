package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/engine/internal/repo/dbstruct"
	"github.com/timickb/sagaflow/lib/db"
	"github.com/timickb/sagaflow/lib/utils"
	"gorm.io/gorm/clause"
)

type instanceRepo struct {
	db *db.Database
}

func NewInstanceRepo(db *db.Database) *instanceRepo {
	return &instanceRepo{
		db: db,
	}
}

// TakeBatch - взять пачку экземпляров с блокировкой для обработки
func (r *instanceRepo) TakeBatch(
	ctx context.Context, batchSize int, lockExpire time.Duration, workerId string,
) ([]*domain.InstanceView, error) {
	var instances []*dbstruct.DBSagaInstance
	lockedTill := time.Now().Add(lockExpire)
	err := r.db.WithTxSupport(ctx).Raw(`
		WITH batch AS (
    		SELECT saga_id
    		FROM saga_instance
    		WHERE status IN ('RUNNING', 'COMPENSATING', 'VERIFYING')
				AND execution_state = 'RUNNABLE'
      			AND next_execution_at <= now()
      			AND (locked_till IS NULL OR locked_till < now())
   	 		ORDER BY next_execution_at, started_at
    		LIMIT $1
    		FOR UPDATE SKIP LOCKED
		)
		UPDATE saga_instance s
		SET locked_till = $2, locked_by = $3, updated_at = now()
		FROM batch
		WHERE s.saga_id = batch.saga_id
		RETURNING s.*;`,
		batchSize, lockedTill, workerId,
	).
		Scan(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("take saga instances batch: %w", err)
	}
	return utils.MapSlice(instances, (*dbstruct.DBSagaInstance).ToDomain), nil
}

// RemoveLock - снять блокировку с экземпляра
func (r *instanceRepo) RemoveLock(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithTxSupport(ctx).
		Model(&dbstruct.DBSagaInstance{SagaId: id}).
		Updates(
			map[string]interface{}{
				"locked_till": (*time.Time)(nil),
				"locked_by":   (*string)(nil),
			},
		).Error
	if err != nil {
		return fmt.Errorf("remove instance lock: %w", err)
	}
	return nil
}

// MakeTransition - перевести экземпляр в следующее состояние
func (r *instanceRepo) MakeTransition(ctx context.Context, dto *domain.InstanceTransitionDto) error {
	updateMap := dbstruct.NewSagaInstanceMakeTransitionUpdateMap(dto)
	err := r.db.WithTxSupport(ctx).
		Model(&dbstruct.DBSagaInstance{SagaId: dto.Id}).
		Updates(updateMap).Error
	if err != nil {
		return fmt.Errorf("make instance transition: %w", err)
	}
	return nil
}

// Create - создать экземпляр, ожидающий начала выполнения
func (r *instanceRepo) Create(ctx context.Context, dto *domain.InstanceStartDto) (uuid.UUID, error) {
	instanceId := uuid.New()
	err := r.db.WithTxSupport(ctx).Create(dbstruct.NewSagaInstance(instanceId, dto)).Error
	if err != nil {
		return uuid.Nil, fmt.Errorf("create db instance: %w", err)
	}
	return instanceId, nil
}

// SetFailed - перевести экземпляр в статус FAILED
func (r *instanceRepo) SetFailed(ctx context.Context, id uuid.UUID, dto *domain.InstanceFailDto) error {
	now := time.Now()
	query := r.db.WithTxSupport(ctx).
		Model(&dbstruct.DBSagaInstance{SagaId: id}).
		Updates(
			map[string]interface{}{
				"status":             domain.InstanceStatusFailed,
				"last_error_code":    dto.ErrCode,
				"last_error_message": dto.ErrMessage,
				"finished_at":        now,
				"updated_at":         now,
			},
		)

	if query.Error != nil {
		return fmt.Errorf("set db instance failed: %w", query.Error)
	}
	if query.RowsAffected == 0 {
		return fmt.Errorf("set db instance failed: no rows affected")
	}
	return nil
}

// GetForEvent - получить конкретный экземпляр с блокировкой
func (r *instanceRepo) GetForEvent(
	ctx context.Context, id uuid.UUID, lockExpire time.Duration, workerId string,
) (*domain.InstanceView, error) {
	now := time.Now()
	lockedTill := now.Add(lockExpire)
	suitableStatuses := []domain.InstanceStatus{
		domain.InstanceStatusRunning,
		domain.InstanceStatusCompleted,
	}

	var instance *dbstruct.DBSagaInstance

	query := r.db.WithTxSupport(ctx).
		Model(&instance).
		Clauses(clause.Returning{
			Columns: []clause.Column{{Name: "*"}},
		}).
		Where("saga_id = ?", id).
		Where("next_execution_at <= now()").
		Where("locked_till IS NULL OR locked_till < now()").
		Where("status IN (?)", suitableStatuses).
		Where("execution_state = ?", domain.InstanceExecutionStateWaitEvent).
		Updates(map[string]interface{}{
			"locked_till": lockedTill,
			"locked_by":   workerId,
			"updated_at":  now,
		})
	if query.Error != nil {
		return nil, fmt.Errorf("get db instance failed: %w", query.Error)
	}
	if query.RowsAffected == 0 {
		return nil, fmt.Errorf("get db instance failed: instance %v not found", id)
	}

	return instance.ToDomain(), nil
}
