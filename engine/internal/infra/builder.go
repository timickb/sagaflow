package infra

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/config"
	"github.com/timickb/sagaflow/engine/internal/engine"
	"github.com/timickb/sagaflow/engine/internal/repo"
	"github.com/timickb/sagaflow/engine/migrations"
	"github.com/timickb/sagaflow/engine/pkg/broker"
	"github.com/timickb/sagaflow/engine/pkg/db"
)

type Builder struct {
	ctx      context.Context
	cfg      *config.Config
	db       *db.Database
	consumer *broker.KafkaStepResultReader
	runner   *engine.Runner
}

func NewBuilder(ctx context.Context, cfg *config.Config) (*Builder, error) {
	b := &Builder{
		ctx: ctx,
		cfg: cfg,
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

	return b, nil
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

func (b *Builder) buildRunner() error {
	b.runner = engine.NewRunner(b.cfg.Runner, repo.NewInstanceRepo(b.db))
	return nil
}

func (b *Builder) Start() error {
	// 1. Запуск асинхронного обработчика инстансов
	if err := b.runner.Run(b.ctx); err != nil {
		return fmt.Errorf("start runner: %w", err)
	}
	// 2. Запуск консьюмера для кафки
	err := b.consumer.Start(b.ctx, func(ctx context.Context, event *broker.SagaStepResultEvent) error {
		log.Info().Msgf(
			"received step result: saga_id=%s step=%s status=%s service=%s",
			event.Ref.SagaId,
			event.Ref.StepName,
			event.Status,
			event.Ref.ServiceName,
		)

		// обработать событие
		return nil
	})
	if err != nil {
		return fmt.Errorf("read kafka event: %w", err)
	}
	return nil
}
