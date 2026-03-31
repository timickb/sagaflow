package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/engine/internal/repo/dbstruct"
	"github.com/timickb/sagaflow/engine/pkg/db"
	"github.com/timickb/sagaflow/engine/pkg/utils"
)

type instanceRepo struct {
	db *db.Database
}

func NewInstanceRepo(db *db.Database) *instanceRepo {
	return &instanceRepo{
		db: db,
	}
}

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
		SET locked_till = now() + $2::interval,
    		locked_by = $3,
    		updated_at = now()
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

func (r *instanceRepo) Create(ctx context.Context, dto *domain.InstanceStartDto) (uuid.UUID, error) {
	instanceId := uuid.New()
	err := r.db.WithTxSupport(ctx).Create(dbstruct.NewSagaInstance(instanceId, dto)).Error
	if err != nil {
		return uuid.Nil, fmt.Errorf("create db instance: %w", err)
	}
	return instanceId, nil
}
