package infra

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/config"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/engine/internal/dsl"
	"github.com/timickb/sagaflow/engine/internal/repo"
	"github.com/timickb/sagaflow/engine/internal/usecase/instance"
	"github.com/timickb/sagaflow/engine/internal/worker"
	"github.com/timickb/sagaflow/engine/migrations"
	"github.com/timickb/sagaflow/engine/pkg/broker"
	"github.com/timickb/sagaflow/engine/pkg/db"
)

type Builder struct {
	ctx             context.Context
	cfg             *config.Config
	db              *db.Database
	consumer        *broker.KafkaStepResultReader
	runner          *worker.Runner
	sagasCache      domain.SagaCache
	instanceUsecase domain.InstanceUsecase
}

func NewBuilder(cfg *config.Config) (*Builder, error) {
	b := &Builder{cfg: cfg}
	b.buildContext()

	if err := b.buildSagaCache(); err != nil {
		return nil, err
	}
	if err := b.buildDB(); err != nil {
		return nil, err
	}
	if err := b.buildRunner(); err != nil {
		return nil, err
	}
	if err := b.buildConsumer(); err != nil {
		return nil, err
	}
	b.instanceUsecase = instance.NewUsecase(repo.NewInstanceRepo(b.db), b.sagasCache)

	return b, nil
}

func (b *Builder) buildContext() {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT)
	defer signal.Stop(sigCh)
	go func() {
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg("OS signal received")
		cancel()
	}()
	b.ctx = ctx
}

func (b *Builder) buildDB() error {
	d, err := db.CreatePostgresConnection(b.ctx, b.cfg.Postgres)
	if err != nil {
		return fmt.Errorf("create postgres connection: %w", err)
	}

	if b.cfg.Postgres.AutoMigrate {
		sqlDb, dErr := d.SqlDB()
		if dErr != nil {
			return fmt.Errorf("get sql db: %w", dErr)
		}
		dErr = migrations.Migrator.Migrate(sqlDb, b.cfg.Postgres.Name)
		if dErr != nil {
			return fmt.Errorf("make migration: %w", dErr)
		}
	}

	b.db = d
	return nil
}

func (b *Builder) buildConsumer() error {
	reader, err := broker.NewKafkaStepResultReader(b.cfg.Kafka)
	if err != nil {
		return fmt.Errorf("create kafka step result reader: %w", err)
	}
	b.consumer = reader
	return nil
}

func (b *Builder) buildSagaCache() error {
	cache, err := dsl.NewCache(b.cfg.Runner.SagasDirPath)
	if err != nil {
		return fmt.Errorf("create saga cache instance: %w", err)
	}
	b.sagasCache = cache
	return nil
}

func (b *Builder) buildRunner() error {
	b.runner = worker.NewRunner(b.cfg.Runner, repo.NewInstanceRepo(b.db), b.sagasCache)
	return nil
}

func (b *Builder) Start() error {
	ctx := b.ctx
	// 1. Запуск асинхронного обработчика инстансов
	if err := b.runner.Run(ctx); err != nil {
		return fmt.Errorf("start runner: %w", err)
	}
	// 2. Запуск консьюмера для кафки
	err := b.consumer.Start(ctx, func(ctx context.Context, event *broker.SagaStepResultEvent) error {
		log.Info().Msgf(
			"received step result: saga_id=%s step=%s status=%s service=%s",
			event.Ref.SagaId,
			event.Ref.StepName,
			event.Status,
			event.Ref.ServiceName,
		)
		if err := b.instanceUsecase.ApplyStepResult(ctx, event); err != nil {
			return fmt.Errorf("apply step result: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("start kafka consumer: %w", err)
	}
	return nil
}
