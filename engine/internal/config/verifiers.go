package config

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/timickb/sagaflow/engine/internal/domain"
)

type DataSource struct {
	Type       domain.VerificationSourceType `yaml:"type" env:"TYPE"`
	TimeoutRaw string                        `yaml:"timeout" env:"TIMEOUT"`
	Timeout    time.Duration                 `yaml:"-"`

	DSN     string            `yaml:"dsn,omitempty" env:"DSN"`
	BaseUrl string            `yaml:"base_url,omitempty" env:"BASE_URL"`
	Headers map[string]string `yaml:"headers,omitempty" env:"HEADERS"`
}

func (d *DataSource) Validate() error {
	timeout, err := time.ParseDuration(d.TimeoutRaw)
	if err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}
	d.Timeout = timeout

	switch d.Type {
	case domain.VerificationSourceTypeClickHouse, domain.VerificationSourceTypePostgres:
		if d.DSN == "" {
			return errors.New("dsn is required")
		}
	case domain.VerificationSourceTypeRest:
		_, parseErr := url.Parse(d.BaseUrl)
		if parseErr != nil {
			return fmt.Errorf("base url is invalid: %w", parseErr)
		}
	default:
		return errors.New("invalid verification source type")
	}
	return nil
}

type VerifiersConfig struct {
	DataSources map[string]DataSource `yaml:"data_sources"`
}

func (c *VerifiersConfig) Validate() error {
	if c.DataSources == nil {
		return errors.New("data_sources is required")
	}
	for _, params := range c.DataSources {
		if err := params.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *VerifiersConfig) GetVerificationSourceParams(name string) (*domain.VerificationSourceParams, error) {
	raw, ok := c.DataSources[name]
	if !ok {
		return nil, fmt.Errorf("no data source registered with name %s", name)
	}
	return &domain.VerificationSourceParams{
		Type:    raw.Type,
		Timeout: raw.Timeout,
		DSN:     raw.DSN,
		BaseUrl: raw.BaseUrl,
		Headers: raw.Headers,
	}, nil
}
