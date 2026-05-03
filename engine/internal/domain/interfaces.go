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

// HandlersConfig - конфигурация обработчиков локальных транзакций
type HandlersConfig interface {
	GetHandlerConnection(serviceName string) (*grpc.ClientConn, bool)
}

// VerificationSourcesConfig - конфигурация подключений к источникам данных
type VerificationSourcesConfig interface {
	GetVerificationSourceParams(name string) (*VerificationSourceParams, error)
}

// === USECASES ===

// InstanceUsecase - бизнес-логика над экземплярами саг
type InstanceUsecase interface {
	Start(ctx context.Context, dto *InstanceStartDto) (uuid.UUID, error)
	GetView(ctx context.Context, sagaId uuid.UUID) (*InstanceView, error)
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
	Terminate(ctx context.Context, id uuid.UUID, dto *InstanceTerminateDto) error
	SetExecutionState(ctx context.Context, id uuid.UUID, state InstanceExecutionState) error
}

// StepRepository - репозиторий над шагами саги
type StepRepository interface {
	GetByInstanceAndName(ctx context.Context, instanceId uuid.UUID, stepName string) (*StepView, bool, error)
	Create(ctx context.Context, dto *StepCreateDto) (*StepView, error)
	Update(ctx context.Context, dto *StepUpdateDto) error
}

// VerificationSource - адаптер для верификации данных в рамках сверочных шагов
type VerificationSource interface {
	Verify(ctx context.Context, req *VerificationRequest) (*VerificationResult, error)
}
