package config

import (
	"errors"
	"fmt"
	"time"
)

// RunnerConfig - конфигурация асинхронного обработчика экземпляров саг
type RunnerConfig struct {
	SagasDirPath       string `yaml:"sagas_dir_path"`
	Hostname           string `yaml:"hostname"`
	WorkersNum         int    `yaml:"workers_num" env:"WORKERS_NUM"`
	BatchSize          int    `yaml:"batch_size" env:"BATCH_SIZE"`
	EmptyBatchDelayRaw string `yaml:"empty_batch_delay" env:"EMPTY_BATCH_DELAY"`
	LockTimeoutRaw     string `yaml:"lock_timeout" env:"LOCK_TIMEOUT"`

	EmptyBatchDelay time.Duration
	LockTimeout     time.Duration
}

func (c *RunnerConfig) Validate() error {
	if c.SagasDirPath == "" {
		return errors.New("sagas_dir_path is required")
	}
	if c.Hostname == "" {
		return errors.New("hostname is required")
	}
	if c.WorkersNum == 0 {
		return errors.New("workers num is required")
	}
	if c.BatchSize == 0 {
		return errors.New("batch size is required")
	}

	emptyBatchDelay, err := time.ParseDuration(c.EmptyBatchDelayRaw)
	if err != nil {
		return fmt.Errorf("invalid batch delay format: %w", err)
	}
	lockTimeout, err := time.ParseDuration(c.LockTimeoutRaw)
	if err != nil {
		return fmt.Errorf("invalid lock timeout format: %w", err)
	}
	c.EmptyBatchDelay = emptyBatchDelay
	c.LockTimeout = lockTimeout
	return nil
}

func (c *RunnerConfig) GetHostname() string {
	return c.Hostname
}
func (c *RunnerConfig) GetWorkersNum() int {
	return c.WorkersNum
}
func (c *RunnerConfig) GetBatchSize() int {
	return c.BatchSize
}
func (c *RunnerConfig) GetEmptyBatchDelay() time.Duration {
	return c.EmptyBatchDelay
}
func (c *RunnerConfig) GetLockTimeout() time.Duration {
	return c.LockTimeout
}
