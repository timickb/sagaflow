package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/timickb/sagaflow/lib/db"
	"github.com/timickb/sagaflow/lib/outbox"
	"github.com/timickb/sagaflow/proto/gen/go/payments"
	sagaflow "github.com/timickb/sagaflow/proto/gen/go/sagaflow"
	"github.com/timickb/sagaflow/services/payments/internal/controller"
	grpcserver "github.com/timickb/sagaflow/services/payments/internal/grpcserver"
	"github.com/timickb/sagaflow/services/payments/internal/payment"
	"github.com/timickb/sagaflow/services/payments/internal/repository"
	"github.com/timickb/sagaflow/services/payments/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	stepHandlerServerAddr = ":50052"
	paymentsServerAddr    = ":50054"
)

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

	// Инициализация PaymentsService gRPC сервера
	paymentsGrpcServer := grpcserver.NewPaymentsServer(orderRepo)
	paymentsLis, err := net.Listen("tcp", paymentsServerAddr)
	if err != nil {
		log.Fatalf("failed to listen for payments server: %v", err)
	}

	paymentsGrpcServerInstance := grpc.NewServer()
	payments.RegisterPaymentsServiceServer(paymentsGrpcServerInstance, paymentsGrpcServer)

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Запуск PaymentsService gRPC сервера
	go func() {
		log.Printf("Starting PaymentsService gRPC server on %s", paymentsServerAddr)
		if err := paymentsGrpcServerInstance.Serve(paymentsLis); err != nil {
			log.Fatalf("failed to serve payments grpc: %v", err)
		}
	}()

	// Запуск StepHandler gRPC сервера
	lis, err := net.Listen("tcp", stepHandlerServerAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	sagaflow.RegisterStepHandlerServiceServer(grpcServer, ctrl)
	reflection.Register(grpcServer) // для grpcurl/grpcui

	go func() {
		log.Printf("Starting StepHandler gRPC server on %s", stepHandlerServerAddr)
		if err = grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-sigCh
	log.Println("Shutting down...")
	grpcServer.GracefulStop()
	paymentsGrpcServerInstance.GracefulStop()
}
