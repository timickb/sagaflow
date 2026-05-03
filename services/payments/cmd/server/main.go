package main

import (
	"context"
	"log"
	"net"

	"github.com/timickb/sagaflow/lib/db"
	"github.com/timickb/sagaflow/lib/outbox"
	sagaflow "github.com/timickb/sagaflow/proto/gen/go/sagaflow"
	"github.com/timickb/sagaflow/services/payments/internal/controller"
	"github.com/timickb/sagaflow/services/payments/internal/payment"
	"github.com/timickb/sagaflow/services/payments/internal/repository"
	"github.com/timickb/sagaflow/services/payments/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const grpcServerAddr = ":50052"

func main() {
	ctx := context.Background()

	// Подключение к БД
	pgDb, err := db.CreatePostgresConnection(ctx, &db.PostgresConfig{
		Host:     "localhost",
		Name:     "payments",
		User:     "postgres",
		Password: "postgres",
		SSLMode:  "disable",
		Port:     5993,
	})
	if err != nil {
		log.Fatalf("create postgres connection: %v", err)
	}

	// Инициализация слоя данных
	orderRepo := repository.NewOrderRepository(pgDb)
	paymentRepo := repository.NewPaymentRepository(pgDb)
	outboxRepo := repository.NewOutboxRepository(outbox.NewRepository(pgDb))
	analyticsOutboxRepo := repository.NewAnalyticsOutboxRepository(pgDb)
	transactor := db.NewTransactor(pgDb)

	// Инициализация платежного провайдера (заглушка)
	paymentProvider := payment.NewStubPaymentProvider()

	// Инициализация сервиса
	svc := service.NewPaymentsService(
		orderRepo,
		paymentRepo,
		outboxRepo,
		analyticsOutboxRepo,
		transactor,
		paymentProvider,
	)

	// Инициализация контроллера
	ctrl := controller.NewStepHandlerController(svc)

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", grpcServerAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	sagaflow.RegisterStepHandlerServiceServer(grpcServer, ctrl)
	reflection.Register(grpcServer) // для grpcurl/grpcui

	log.Printf("Starting gRPC server on %s", grpcServerAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
