package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// RunnerConfig - конфигурация обработчика экземпляров
type RunnerConfig interface {
	GetHostname() string
	GetWorkersNum() int
	GetBatchSize() int
	GetEmptyBatchDelay() time.Duration
	GetLockTimeout() time.Duration
}

// === USECASES ===

// InstanceUsecase - бизнес-логика над экземплярами саг
type InstanceUsecase interface {
	Start(ctx context.Context, dto *InstanceStartDto) (uuid.UUID, error)
	GetInfo(ctx context.Context, sagaId uuid.UUID) (*InstanceView, error)
	GetFeedCount(ctx context.Context, dto *GetFeedDto) (int64, error)
	GetFeed(ctx context.Context, dto *GetFeedDto) (*InstancesFeed, error)
	GetFeedNext(ctx context.Context, paginationToken string) (*InstancesFeed, error)
}

// === REPOSITORIES ===

type Transactor interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// SagaDefinitionCache - кэш моделей (описаний) саг
type SagaDefinitionCache interface {
	GetSagaDefinition(header SagaDefinitionHeader) (*SagaDefinition, bool)
}

// HandlerCache - кэш соединений для хэндлеров
type HandlerCache interface {
	GetHandlerClient(handler *Handler) (*grpc.ClientConn, bool)
}

// InstanceRepository - репозиторий наг экземплярами саг
type InstanceRepository interface {
	TakeBatch(
		ctx context.Context,
		batchSize int,
		lockExpire time.Duration,
		workerId string,
	) ([]*InstanceView, error)
	GetForEvent(
		ctx context.Context,
		id uuid.UUID,
		lockExpire time.Duration,
		workerId string,
	) (*InstanceView, error)
	MakeTransition(ctx context.Context, dto *InstanceTransitionDto) error
	RemoveLock(ctx context.Context, id uuid.UUID) error
	Create(ctx context.Context, dto *InstanceStartDto) (uuid.UUID, error)
	SetFailed(ctx context.Context, id uuid.UUID, dto *InstanceFailDto) error
}

// StepRepository - репозиторий над шагами саги
type StepRepository interface {
	GetByInstanceAndName(ctx context.Context, instanceId uuid.UUID, stepName string) (*StepView, bool, error)
	Create(ctx context.Context, dto *StepCreateDto) error
	Update(ctx context.Context, dto *StepUpdateDto) error
}
