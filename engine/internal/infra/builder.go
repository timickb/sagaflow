package infra

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/adapter/step_handler"
	"github.com/timickb/sagaflow/engine/internal/api"
	"github.com/timickb/sagaflow/engine/internal/config"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/engine/internal/dsl"
	"github.com/timickb/sagaflow/engine/internal/repo"
	"github.com/timickb/sagaflow/engine/internal/usecase/instance"
	"github.com/timickb/sagaflow/engine/internal/worker"
	"github.com/timickb/sagaflow/engine/migrations"
	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/db"
	pb "github.com/timickb/sagaflow/proto/gen/go/sagaflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Builder struct {
	ctx                context.Context
	cfg                *config.Config
	db                 *db.Database
	consumer           *broker.KafkaStepResultReader
	publisher          *broker.KafkaStepResultWriter
	runner             *worker.Runner
	server             *grpc.Server
	sagasCache         domain.SagaDefinitionCache
	instanceUsecase    domain.InstanceUsecase
	stepHandlerAdapter domain.StepHandler
}

func NewBuilder(cfg *config.Config) (*Builder, error) {
	b := &Builder{cfg: cfg}
	b.buildContext()

	// кэш определений саг
	if err := b.buildSagaDefinitionCache(); err != nil {
		return nil, err
	}
	// подключение к базе данных
	if err := b.buildDB(); err != nil {
		return nil, err
	}
	// бизнес-логика над экземплярами саг
	if err := b.buildInstanceUsecase(); err != nil {
		return nil, err
	}
	// адаптер для вызова обработчиков шагов
	if err := b.buildStepHandlerAdapter(); err != nil {
		return nil, err
	}
	// publisher для kafka
	if err := b.buildPublisher(); err != nil {
		return nil, err
	}
	// consumer для kafka
	if err := b.buildConsumer(); err != nil {
		return nil, err
	}
	// асинхронный обработчик экземпляров
	if err := b.buildRunner(); err != nil {
		return nil, err
	}
	// api оркестратора
	if err := b.buildServer(); err != nil {
		return nil, err
	}

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

func (b *Builder) buildInstanceUsecase() error {
	b.instanceUsecase = instance.NewUsecase(
		repo.NewInstanceRepo(b.db),
		repo.NewStepRepo(b.db),
		db.NewTransactor(b.db),
		b.sagasCache,
	)
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

func (b *Builder) buildPublisher() error {
	writer, err := broker.NewKafkaStepResultWriter(b.cfg.Kafka)
	if err != nil {
		return fmt.Errorf("create kafka step result writer: %w", err)
	}
	b.publisher = writer
	return nil
}

func (b *Builder) buildSagaDefinitionCache() error {
	cache, err := dsl.NewCache(b.cfg.Runner.SagasDirPath)
	if err != nil {
		return fmt.Errorf("create saga cache instance: %w", err)
	}
	b.sagasCache = cache
	return nil
}

func (b *Builder) buildRunner() error {
	b.runner = worker.NewRunner(
		b.cfg.Runner,
		b.cfg.Handlers,
		b.cfg.Verifiers,
		repo.NewInstanceRepo(b.db),
		repo.NewStepRepo(b.db),
		db.NewTransactor(b.db),
		b.sagasCache,
		b.stepHandlerAdapter,
		b.publisher,
	)
	return nil
}

func (b *Builder) buildServer() error {
	server := grpc.NewServer()
	sagaflowServer := api.NewSagaflowServer(b.instanceUsecase)
	pb.RegisterSagaflowServiceServer(server, sagaflowServer)
	reflection.Register(server)
	b.server = server
	return nil
}

func (b *Builder) buildStepHandlerAdapter() error {
	adapter, err := step_handler.NewAdapter(b.cfg.Handlers)
	if err != nil {
		return fmt.Errorf("create step handler adapter: %w", err)
	}
	b.stepHandlerAdapter = adapter
	return nil
}

func (b *Builder) runServer() error {
	addr := fmt.Sprintf(":%d", b.cfg.Api.Port)
	log.Info().Int("port", b.cfg.Api.Port).Msg("Starting gRPC server")

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	if err = b.server.Serve(lis); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func (b *Builder) Start() *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(3)
	ctx := b.ctx
	// 1. Запуск асинхронного обработчика инстансов
	go func() {
		defer wg.Done()
		if err := b.runner.Run(ctx); err != nil {
			log.Fatal().Err(err).Msg("Runner start failed")
		}
	}()

	// 2. Запуск консьюмера для кафки
	go func() {
		defer wg.Done()
		err := b.consumer.Start(ctx, func(ctx context.Context, event *broker.SagaStepResultEvent) error {
			log.Info().Msgf(
				"received step result: saga_id=%s step=%s status=%s service=%s",
				event.Ref.SagaId,
				event.Ref.StepName,
				event.Status,
				event.Ref.ServiceName,
			)
			if err := b.runner.ApplyStepResult(ctx, event); err != nil {
				return fmt.Errorf("apply step result: %w", err)
			}
			return nil
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Consumer start failed")
		}
	}()

	// 3. Запуск gRPC сервера
	go func() {
		defer wg.Done()
		if err := b.runServer(); err != nil {
			log.Fatal().Err(err).Msg("gRPC server start failed")
		}
	}()

	return wg
}
