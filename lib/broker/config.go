package broker

import (
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	DefaultKafkaDialTimeout    = 10 * time.Second
	DefaultKafkaReadTimeout    = 10 * time.Second
	DefaultKafkaCommitInterval = 1 * time.Second
	DefaultKafkaMaxBytes       = 10e6
)

type KafkaConfig struct {
	Brokers  []string `yaml:"brokers"`
	GroupId  string   `yaml:"group_id"`
	ClientId string   `yaml:"client_id"`
	// StepResultTopic - название топика, из которого читаются события коммита локальных транзакций
	StepResultTopic   string `yaml:"step_result_topic"`
	DialTimeoutRaw    string `yaml:"dial_timeout"`
	ReadTimeoutRaw    string `yaml:"read_timeout"`
	CommitIntervalRaw string `yaml:"commit_interval"`
	// MaxBytes - максимальный размер сообщения
	MaxBytes    int   `yaml:"max_bytes"`
	StartOffset int64 `yaml:"start_offset"`

	DialTimeout    time.Duration
	ReadTimeout    time.Duration
	CommitInterval time.Duration
}

func (c *KafkaConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("kafka brokers are required")
	}
	if c.GroupId == "" {
		return errors.New("kafka group_id is required")
	}
	if c.StepResultTopic == "" {
		return errors.New("kafka topic is required")
	}

	dialTimeout, err := time.ParseDuration(c.DialTimeoutRaw)
	if err != nil {
		return fmt.Errorf("invalid dial_timeout: %w", err)
	}
	c.DialTimeout = dialTimeout

	readTimeout, err := time.ParseDuration(c.ReadTimeoutRaw)
	if err != nil {
		return fmt.Errorf("invalid read_timeout: %w", err)
	}
	c.ReadTimeout = readTimeout

	commitInterval, err := time.ParseDuration(c.CommitIntervalRaw)
	if err != nil {
		return fmt.Errorf("invalid commit_interval: %w", err)
	}
	c.CommitInterval = commitInterval

	return nil
}

func (c *KafkaConfig) withDefaults() *KafkaConfig {
	if c.DialTimeout <= 0 {
		c.DialTimeout = DefaultKafkaDialTimeout
	}
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = DefaultKafkaReadTimeout
	}
	if c.CommitInterval <= 0 {
		c.CommitInterval = DefaultKafkaCommitInterval
	}
	if c.MaxBytes <= 0 {
		c.MaxBytes = DefaultKafkaMaxBytes
	}
	if c.StartOffset != kafka.FirstOffset && c.StartOffset != kafka.LastOffset {
		c.StartOffset = kafka.FirstOffset
	}
	return c
}
