package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// HandlersConfig - конфигурация подключений к сервисам-обработчикам шагов саг
type HandlersConfig struct {
	CallTimeoutRaw string `yaml:"call_timeout"`
	CallTimeout    time.Duration
	TLS            bool     `yaml:"tls"`
	Endpoints      []string `yaml:"endpoints"`

	parsedEndpoints map[string]string
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

func (c *HandlersConfig) GetEndpoints() map[string]string {
	return c.parsedEndpoints
}

func (c *HandlersConfig) GetTLS() bool {
	return c.TLS
}

func (c *HandlersConfig) parseEndpoints() error {
	c.parsedEndpoints = make(map[string]string)
	for _, endpoint := range c.Endpoints {
		parts := strings.Split(endpoint, ":")
		if len(parts) != 3 {
			return fmt.Errorf("invalid handler: %s", endpoint)
		}
		if _, err := strconv.Atoi(parts[2]); err != nil {
			return fmt.Errorf("invalid port in handler: %s", endpoint)
		}
		c.parsedEndpoints[parts[0]] = fmt.Sprintf("%s:%s", parts[1], parts[2])
	}
	return nil
}
