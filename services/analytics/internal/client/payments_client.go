package client

import (
	"context"

	"github.com/timickb/sagaflow/proto/gen/go/payments"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentsClient struct {
	conn   *grpc.ClientConn
	client payments.PaymentsServiceClient
}

func NewPaymentsClient(addr string) (*PaymentsClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &PaymentsClient{
		conn:   conn,
		client: payments.NewPaymentsServiceClient(conn),
	}, nil
}

func (c *PaymentsClient) GetOrderById(ctx context.Context, orderId string) (*payments.GetOrderByIdResponse, error) {
	return c.client.GetOrderById(ctx, &payments.GetOrderByIdRequest{
		OrderId: orderId,
	})
}

func (c *PaymentsClient) Close() error {
	return c.conn.Close()
}
