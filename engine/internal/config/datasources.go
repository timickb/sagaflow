package config

import (
	"time"
)

type DataSourcesConfig struct {
	CallTimeoutRaw string `yaml:"call_timeout"`
	CallTimeout    time.Duration
	DataSources    []string `yaml:"data_sources"`

	connections map[string]string
}
