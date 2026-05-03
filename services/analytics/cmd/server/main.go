package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	sagaflow "github.com/timickb/sagaflow/proto/gen/go/sagaflow"
	"github.com/timickb/sagaflow/services/analytics/internal/clickhouse"
	"github.com/timickb/sagaflow/services/analytics/internal/consumer"
	"github.com/timickb/sagaflow/services/analytics/internal/controller"
	"github.com/timickb/sagaflow/services/analytics/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	grpcServerAddr = ":50053"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Конфигурация ClickHouse
	chRepo, err := clickhouse.NewRepository(&clickhouse.Config{
		Addresses: []string{"localhost:9000"},
		Database:  "analytics",
		Username:  "default",
		Password:  "",
	})
	if err != nil {
		log.Fatalf("create clickhouse repository: %v", err)
	}
	defer chRepo.Close()

	// Инициализация сервиса
	svc := service.NewAnalyticsService(chRepo)

	// Kafka консьюмер
	kafkaCfg := &consumer.KafkaConsumerConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "analytics-service",
		Topic:   consumer.OrdersEventsTopic,
	}
	kafkaConsumer := consumer.NewKafkaConsumer(kafkaCfg)
	defer kafkaConsumer.Stop()

	// Запуск Kafka консьюмера в горутине
	go func() {
		if err := kafkaConsumer.Start(ctx, svc); err != nil {
			log.Printf("kafka consumer error: %v", err)
		}
	}()

	// gRPC сервер
	ctrl := controller.NewStepHandlerController(svc)

	lis, err := net.Listen("tcp", grpcServerAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	sagaflow.RegisterStepHandlerServiceServer(grpcServer, ctrl)
	reflection.Register(grpcServer)

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("starting gRPC server on %s", grpcServerAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-sigCh
	log.Println("Shutting down...")
	cancel()
	grpcServer.GracefulStop()
}
