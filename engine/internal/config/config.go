package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/timickb/sagaflow/engine/pkg/broker"
	"github.com/timickb/sagaflow/engine/pkg/db"
	"gopkg.in/yaml.v3"
)

// Config - конфигурация сервиса
type Config struct {
	Postgres   *db.PostgresConfig  `yaml:"postgres" env:"POSTGRES"`
	Kafka      *broker.KafkaConfig `yaml:"kafka" env:"KAFKA"`
	Runner     *RunnerConfig       `yaml:"runner" env:"RUNNER"`
	Backoffice *BackofficeConfig   `yaml:"backoffice" env:"BACKOFFICE"`
}

// NewFromFile - создать конфиг из YAML файла
func NewFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := Config{}
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal yml config: %w", err)
	}
	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Postgres == nil {
		return errors.New("postgres configuration is required")
	}
	if c.Kafka == nil {
		return errors.New("kafka configuration is required")
	}
	if c.Runner == nil {
		return errors.New("runner configuration is required")
	}
	if c.Backoffice == nil {
		return errors.New("backoffice configuration is required")
	}
	return nil
}
