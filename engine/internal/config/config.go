package config

import (
	"fmt"
	"os"

	"github.com/timickb/sagaflow/lib/broker"
	"github.com/timickb/sagaflow/lib/db"
	"gopkg.in/yaml.v3"
)

type validatable interface {
	Validate() error
}

// Config - конфигурация сервиса
type Config struct {
	Postgres *db.PostgresConfig  `yaml:"postgres" env:"POSTGRES"`
	Kafka    *broker.KafkaConfig `yaml:"kafka" env:"KAFKA"`
	Runner   *RunnerConfig       `yaml:"runner" env:"RUNNER"`
	Api      *APIConfig          `yaml:"api" env:"API"`
	Handlers *HandlersConfig     `yaml:"handlers" env:"HANDLERS"`
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
	checks := []struct {
		name string
		cfg  validatable
	}{
		{"postgres", c.Postgres},
		{"kafka", c.Kafka},
		{"runner", c.Runner},
		{"backoffice", c.Api},
		{"handlers", c.Handlers},
	}

	for _, check := range checks {
		if check.cfg == nil {
			return fmt.Errorf("%s configuration is required", check.name)
		}
		if err := check.cfg.Validate(); err != nil {
			return fmt.Errorf("%s: %w", check.name, err)
		}
	}

	return nil
}
