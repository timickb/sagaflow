package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/engine/internal/repo/dbstruct"
	"github.com/timickb/sagaflow/lib/db"
	"github.com/timickb/sagaflow/lib/utils"
	"gorm.io/gorm"
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
    		WHERE status IN ('PENDING', 'RUNNING', 'COMPENSATING', 'VERIFYING')
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
	return utils.MapSliceOnError(instances, (*dbstruct.DBSagaInstance).ToDomain)
}

// TakeExpiredBatch - получить пачку экземпляров, не дождавшихся вовремя события.
func (r *instanceRepo) TakeExpiredBatch(
	ctx context.Context, batchSize int, lockExpire time.Duration, workerId string,
) ([]*domain.InstanceView, error) {
	var instances []*dbstruct.DBSagaInstance
	lockedTill := time.Now().Add(lockExpire)
	err := r.db.WithTxSupport(ctx).Raw(`
		WITH batch AS (
    		SELECT saga_id
    		FROM saga_instance
    		WHERE execution_state = 'WAITING_EVENT'
      			AND event_timeout_at < now()
      			AND (locked_till IS NULL OR locked_till < now())
   	 		ORDER BY event_timeout_at, started_at
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
		return nil, fmt.Errorf("take expired instances batch: %w", err)
	}
	return utils.MapSliceOnError(instances, (*dbstruct.DBSagaInstance).ToDomain)
}

// RemoveLock - снять блокировку с экземпляра
func (r *instanceRepo) RemoveLock(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithTxSupport(ctx).
		Model(&dbstruct.DBSagaInstance{}).
		Where("saga_id = ?", id).
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
		Model(&dbstruct.DBSagaInstance{}).
		Where("saga_id = ?", dto.Id).
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

// Terminate - перевести экземпляр в терминальный статус
func (r *instanceRepo) Terminate(ctx context.Context, id uuid.UUID, dto *domain.InstanceTerminateDto) error {
	if !dto.Status.IsTerminal() {
		return errors.New("set db instance terminated: only terminal statuses are allowed")
	}
	now := time.Now()
	query := r.db.WithTxSupport(ctx).
		Model(&dbstruct.DBSagaInstance{}).
		Where("saga_id = ?", id).
		Updates(
			map[string]interface{}{
				"status":             dto.Status,
				"last_error_code":    dto.ErrCode,
				"last_error_message": dto.ErrMessage,
				"finished_at":        now,
				"updated_at":         now,
			},
		)

	if query.Error != nil {
		return fmt.Errorf("set db instance terminated: %w", query.Error)
	}
	if query.RowsAffected == 0 {
		return fmt.Errorf("set db instance terminated: no rows affected")
	}
	return nil
}

// GetForEvent - получить конкретный экземпляр с блокировкой
func (r *instanceRepo) GetForEvent(ctx context.Context, id uuid.UUID) (*domain.InstanceView, error) {
	suitableStatuses := []domain.InstanceStatus{
		domain.InstanceStatusPending,
		domain.InstanceStatusRunning,
		domain.InstanceStatusCompensating,
		domain.InstanceStatusVerifying,
	}
	var instance dbstruct.DBSagaInstance
	query := r.db.WithTxSupport(ctx).
		//Clauses(clause.Locking{
		//	Strength: "UPDATE",
		//	Options:  "SKIP LOCKED",
		//}).
		Where("saga_id = ?", id).
		Where("status IN ?", suitableStatuses).
		Where("execution_state = ?", domain.InstanceExecutionStateWaitingEvent).
		First(&instance)
	if query.Error != nil {
		if errors.Is(query.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get db instance failed: %w", query.Error)
	}
	return instance.ToDomain()
}

// SetWaitingEvent - перевести в состояние ожидания события
func (r *instanceRepo) SetWaitingEvent(ctx context.Context, id uuid.UUID, timeout time.Duration) error {
	now := time.Now()
	query := r.db.WithTxSupport(ctx).
		Model(&dbstruct.DBSagaInstance{}).
		Where("saga_id = ?", id).
		Updates(
			map[string]interface{}{
				"execution_state":  domain.InstanceExecutionStateWaitingEvent,
				"event_timeout_at": now.Add(timeout),
				"updated_at":       now,
			},
		)

	if query.Error != nil {
		return fmt.Errorf("set db instance waiting handler: %w", query.Error)
	}
	if query.RowsAffected == 0 {
		return fmt.Errorf("set db instance waiting handler: no rows affected")
	}
	return nil
}
