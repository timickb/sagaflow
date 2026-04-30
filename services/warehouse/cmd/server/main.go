package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/lib/db"
	"github.com/timickb/sagaflow/lib/outbox"
	"github.com/timickb/sagaflow/proto/gen/go/sagaflow"
	pb "github.com/timickb/sagaflow/proto/gen/go/warehouse"
	"github.com/timickb/sagaflow/services/warehouse/internal/repository"
	"github.com/timickb/sagaflow/services/warehouse/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type warehouseServer struct {
	pb.UnimplementedWarehouseServiceServer
	service *service.WarehouseService
}

func newWarehouseServer(svc *service.WarehouseService) *warehouseServer {
	return &warehouseServer{
		service: svc,
	}
}

func (s *warehouseServer) Reserve(ctx context.Context, req *pb.ReserveRequest) (*pb.ReserveResponse, error) {
	// Parse meta
	scenarioInstanceID, err := uuid.Parse(req.Meta.SagaId)
	if err != nil {
		return nil, fmt.Errorf("invalid saga_id in meta: %w", err)
	}

	stepID := req.Meta.StepId
	if stepID == "" {
		stepID = "Reserve"
	}

	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, fmt.Errorf("invalid order_id: %w", err)
	}

	// Convert items
	var items []service.ReserveItem
	for _, item := range req.Items {
		items = append(items, service.ReserveItem{
			SKU:      item.Sku,
			Quantity: int(item.Quantity),
		})
	}

	// Call service
	result, err := s.service.Reserve(ctx, orderID, scenarioInstanceID, stepID, items)
	if err != nil {
		return nil, fmt.Errorf("reserve failed: %w", err)
	}

	// Convert movement IDs to strings
	var movementIDs []string
	for _, id := range result.MovementIDs {
		movementIDs = append(movementIDs, id.String())
	}

	return &pb.ReserveResponse{
		ReservationId: result.ReservationID.String(),
		MovementIds:   movementIDs,
	}, nil
}

func (s *warehouseServer) Release(ctx context.Context, req *pb.ReleaseRequest) (*pb.ReleaseResponse, error) {
	// Parse meta
	scenarioInstanceID, err := uuid.Parse(req.Meta.SagaId)
	if err != nil {
		return nil, fmt.Errorf("invalid saga_id in meta: %w", err)
	}

	stepID := req.Meta.StepId
	if stepID == "" {
		stepID = "Release"
	}

	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, fmt.Errorf("invalid order_id: %w", err)
	}

	// Call service
	result, err := s.service.Release(ctx, orderID, scenarioInstanceID, stepID)
	if err != nil {
		return nil, fmt.Errorf("release failed: %w", err)
	}

	// Convert movement IDs to strings
	var movementIDs []string
	for _, id := range result.MovementIDs {
		movementIDs = append(movementIDs, id.String())
	}

	return &pb.ReleaseResponse{
		MovementIds: movementIDs,
	}, nil
}

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

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWarehouseServiceServer(s, newWarehouseServer(svc))
	sagaflow.
		reflection.Register(s)

	log.Printf("Warehouse service listening on :50051")

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
