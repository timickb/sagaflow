package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
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
	GetEndpoints() map[string]string
	GetTLS() bool
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
	TakeExpiredBatch(
		ctx context.Context,
		batchSize int,
		lockExpire time.Duration,
		workerId string,
	) ([]*InstanceView, error)
	GetForEvent(ctx context.Context, id uuid.UUID) (*InstanceView, error)
	MakeTransition(ctx context.Context, dto *InstanceTransitionDto) error
	RemoveLock(ctx context.Context, id uuid.UUID) error
	Create(ctx context.Context, dto *InstanceStartDto) (uuid.UUID, error)
	Terminate(ctx context.Context, id uuid.UUID, dto *InstanceTerminateDto) error
	SetWaitingEvent(ctx context.Context, id uuid.UUID, timeout time.Duration) error
}

// StepRepository - репозиторий над шагами саги
type StepRepository interface {
	GetByInstanceAndName(ctx context.Context, instanceId uuid.UUID, stepName string) (*StepView, bool, error)
	Upsert(ctx context.Context, dto *StepCreateDto) (*StepView, error)
	Update(ctx context.Context, dto *StepUpdateDto) error
}

// === ADAPTERS ===

// VerificationSource - адаптер для верификации данных в рамках сверочных шагов
type VerificationSource interface {
	Verify(ctx context.Context, req *VerificationRequest) (*VerificationResult, error)
}

// StepHandler - адаптер для вызова обработчиков шагов
type StepHandler interface {
	Call(ctx context.Context, dto *CallHandlerRequest) (*CallHandlerResult, error)
}
