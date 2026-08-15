package grpcserver

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timickb/sagaflow/proto/gen/go/payments"
	"github.com/timickb/sagaflow/services/payments/internal/repository"
)

type PaymentsServer struct {
	payments.UnimplementedPaymentsServiceServer
	orderRepo *repository.OrderRepository
}

func NewPaymentsServer(orderRepo *repository.OrderRepository) *PaymentsServer {
	return &PaymentsServer{orderRepo: orderRepo}
}

func (s *PaymentsServer) GetOrderById(ctx context.Context, req *payments.GetOrderByIdRequest) (*payments.GetOrderByIdResponse, error) {
	orderIdStr := req.GetOrderId()
	if orderIdStr == "" {
		return nil, fmt.Errorf("order_id is required")
	}

	orderId, err := uuid.Parse(orderIdStr)
	if err != nil {
		return nil, fmt.Errorf("invalid order_id format: %w", err)
	}

	order, err := s.orderRepo.GetByID(ctx, orderId)
	if err != nil {
		return nil, fmt.Errorf("get order by id: %w", err)
	}

	if order == nil {
		return nil, fmt.Errorf("order not found: %s", orderId)
	}

	return &payments.GetOrderByIdResponse{
		OrderId:     order.ID.String(),
		UserId:      order.UserID.String(),
		Status:      string(order.Status),
		TotalAmount: float32(order.TotalAmount),
		Currency:    order.Currency,
		CreatedAt:   order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   order.UpdatedAt.Format(time.RFC3339),
		Version:     int32(order.Version),
	}, nil
}
