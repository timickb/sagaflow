package config

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// HandlersConfig - конфигурация подключений к сервисам-обработчикам шагов саг
type HandlersConfig struct {
	CallTimeoutRaw string `yaml:"call_timeout"`
	CallTimeout    time.Duration
	Endpoints      []string `yaml:"endpoints"`

	connections []*grpc.ClientConn
}

func (c *HandlersConfig) Validate() error {
	callTimeout, err := time.ParseDuration(c.CallTimeoutRaw)
	if err != nil {
		return fmt.Errorf("invalid call_timeout: %w", err)
	}
	c.CallTimeout = callTimeout

	if err = c.parseEndpoints(); err != nil {
		return fmt.Errorf("invalid endpoints: %w", err)
	}
	return nil
}

func (c *HandlersConfig) parseEndpoints() error {
	c.connections = make([]*grpc.ClientConn, 0, len(c.Endpoints))
	for _, endpoint := range c.Endpoints {
		parts := strings.Split(endpoint, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid handler: %s", endpoint)
		}
		if _, err := strconv.Atoi(parts[1]); err != nil {
			return fmt.Errorf("invalid port in handler: %s", endpoint)
		}
		conn, err := c.createConnection(endpoint)
		if err != nil {
			return fmt.Errorf("create connection for endpoint %s: %w", endpoint, err)
		}
		c.connections = append(c.connections, conn)
	}
	return nil
}

func (c *HandlersConfig) createConnection(endpoint string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
	}
	conn, err := grpc.NewClient(endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc connection: %w", err)
	}
	return conn, nil
}
