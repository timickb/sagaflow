package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/lib/db"
	"github.com/timickb/sagaflow/lib/outbox"
	sagaflow "github.com/timickb/sagaflow/proto/gen/go/sagaflow"
	"github.com/timickb/sagaflow/services/warehouse/internal/repository"
	"github.com/timickb/sagaflow/services/warehouse/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	grpcServerAddr = ":50051"
)

// StepHandlerController реализует StepHandlerServiceServer
type StepHandlerController struct {
	sagaflow.UnimplementedStepHandlerServiceServer
	svc *service.WarehouseService
}

func NewStepHandlerController(svc *service.WarehouseService) *StepHandlerController {
	return &StepHandlerController{svc: svc}
}

// Handle обрабатывает запросы саги
func (c *StepHandlerController) Handle(ctx context.Context, req *sagaflow.HandleRequest) (*sagaflow.HandleResponse, error) {
	if req.Meta == nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString("meta is required"),
		}, nil
	}

	action := req.Meta.GetAction()
	switch action {
	case "Reserve":
		return c.handleReserve(ctx, req)
	case "Release":
		return c.handleRelease(ctx, req)
	case "Confirm":
		return c.handleConfirm(ctx, req)
	default:
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("unknown action: %s", action)),
		}, nil
	}
}

type reservePayload struct {
	OrderID string                `json:"order_id"`
	Items   []service.ReserveItem `json:"items"`
}

type releasePayload struct {
	OrderID string `json:"order_id"`
}

type confirmPayload struct {
	OrderID string `json:"order_id"`
}

func (c *StepHandlerController) handleReserve(ctx context.Context, req *sagaflow.HandleRequest) (*sagaflow.HandleResponse, error) {
	var payload reservePayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("parse payload: %v", err)),
		}, nil
	}

	orderID, err := uuid.Parse(payload.OrderID)
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("parse order_id: %v", err)),
		}, nil
	}

	sagaID, err := uuid.Parse(req.Meta.GetSagaId())
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("parse saga_id: %v", err)),
		}, nil
	}

	_, err = c.svc.Reserve(ctx, orderID, sagaID, req.Meta.GetStepId(), payload.Items)
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("reserve failed: %v", err)),
		}, fmt.Errorf("reserve failed: %w", err)
	}

	return &sagaflow.HandleResponse{Success: true}, nil
}

func (c *StepHandlerController) handleRelease(ctx context.Context, req *sagaflow.HandleRequest) (*sagaflow.HandleResponse, error) {
	var payload releasePayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("parse payload: %v", err)),
		}, nil
	}

	orderID, err := uuid.Parse(payload.OrderID)
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("parse order_id: %v", err)),
		}, nil
	}

	sagaID, err := uuid.Parse(req.Meta.GetSagaId())
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("parse saga_id: %v", err)),
		}, nil
	}

	_, err = c.svc.Release(ctx, orderID, sagaID, req.Meta.GetStepId())
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("release failed: %v", err)),
		}, nil
	}

	return &sagaflow.HandleResponse{Success: true}, nil
}

func (c *StepHandlerController) handleConfirm(ctx context.Context, req *sagaflow.HandleRequest) (*sagaflow.HandleResponse, error) {
	var payload confirmPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("parse payload: %v", err)),
		}, nil
	}

	orderID, err := uuid.Parse(payload.OrderID)
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("parse order_id: %v", err)),
		}, nil
	}

	sagaID, err := uuid.Parse(req.Meta.GetSagaId())
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("parse saga_id: %v", err)),
		}, nil
	}

	err = c.svc.Confirm(ctx, orderID, sagaID, req.Meta.GetStepId())
	if err != nil {
		return &sagaflow.HandleResponse{
			Success: false,
			Error:   ptrString(fmt.Sprintf("confirm failed: %v", err)),
		}, nil
	}

	return &sagaflow.HandleResponse{Success: true}, nil
}

func ptrString(s string) *string {
	return &s
}

// Ensure StepHandlerController implements StepHandlerServiceServer
var _ sagaflow.StepHandlerServiceServer = (*StepHandlerController)(nil)

// Sentinel errors for validation
var (
	ErrUnknownAction = errors.New("unknown action")
)

func main() {
	ctx := context.Background()
	pgDb, err := db.CreatePostgresConnection(ctx, &db.PostgresConfig{
		Host:     "localhost",
		Name:     "warehouse",
		User:     "postgres",
		Password: "postgres",
		SSLMode:  "disable",
		Port:     5995,
	})
	if err != nil {
		log.Fatalf("create postgres connection: %v", err)
	}

	svc := service.NewWarehouseService(
		repository.NewWarehouseRepository(pgDb),
		outbox.NewRepository(pgDb),
		db.NewTransactor(pgDb),
	)

	ctrl := NewStepHandlerController(svc)

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", grpcServerAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	sagaflow.RegisterStepHandlerServiceServer(grpcServer, ctrl)
	reflection.Register(grpcServer)

	log.Printf("starting gRPC server on %s", grpcServerAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
