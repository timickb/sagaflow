package step_handler

import (
	"crypto/tls"
	"fmt"

	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/proto/gen/go/sagaflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Adapter struct {
	clients map[string]sagaflow.StepHandlerServiceClient
}

func NewAdapter(cfg domain.HandlersConfig) (*Adapter, error) {
	endpoints := cfg.GetEndpoints()
	tlsEnabled := cfg.GetTLS()

	clients := make(map[string]sagaflow.StepHandlerServiceClient, len(endpoints))
	for serviceName, endpoint := range endpoints {
		conn, err := createConnection(endpoint, tlsEnabled)
		if err != nil {
			return nil, fmt.Errorf("failed to create connection for service %s: %w", serviceName, err)
		}
		clients[serviceName] = sagaflow.NewStepHandlerServiceClient(conn)
	}

	return &Adapter{clients: clients}, nil
}

func createConnection(endpoint string, tlsEnabled bool) (*grpc.ClientConn, error) {
	cred := insecure.NewCredentials()
	if tlsEnabled {
		cred = credentials.NewTLS(&tls.Config{})
	}
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(cred))
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc connection: %w", err)
	}
	return conn, nil
}
