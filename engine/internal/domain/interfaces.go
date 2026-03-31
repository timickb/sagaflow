package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/engine/pkg/broker"
)

type Config interface {
	GetHostname() string
	GetWorkersNum() int
	GetBatchSize() int
	GetEmptyBatchDelay() time.Duration
	GetLockTimeout() time.Duration
}

// SagaCache - кэш моделей (описаний) саг
type SagaCache interface {
	GetSagaView(header SagaHeader) (*SagaView, bool)
}

// === USECASES ===

// InstanceUsecase - бизнес-логика над экземплярами саг
type InstanceUsecase interface {
	Start(ctx context.Context, dto *InstanceStartDto) (uuid.UUID, error)
	GetInfo(ctx context.Context, sagaId uuid.UUID) (*InstanceView, error)
	GetFeedCount(ctx context.Context, dto *GetFeedDto) (int64, error)
	GetFeed(ctx context.Context, dto *GetFeedDto) (*InstancesFeed, error)
	GetFeedNext(ctx context.Context, paginationToken string) (*InstancesFeed, error)
	ApplyStepResult(ctx context.Context, event *broker.SagaStepResultEvent) error
}

// === REPOSITORIES ===

// InstanceRepository - репозиторий наг экземплярами саг
type InstanceRepository interface {
	TakeBatch(
		ctx context.Context,
		batchSize int,
		lockExpire time.Duration,
		workerId string,
	) ([]*InstanceView, error)
	Create(ctx context.Context, dto *InstanceStartDto) (uuid.UUID, error)
}

// StepRepository - репозиторий над шагами саги
type StepRepository interface {
	GetByInstanceAndName(ctx context.Context, instanceId uuid.UUID, stepName string) (*StepView, bool, error)
}
