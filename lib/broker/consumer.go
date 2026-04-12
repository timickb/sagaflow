package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type KafkaStepResultReader struct {
	reader *kafka.Reader
}

// NewKafkaStepResultReader - создать консьюмер для чтения из топика
func NewKafkaStepResultReader(cfg *KafkaConfig) (*KafkaStepResultReader, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate kafka config: %w", err)
	}
	cfg = cfg.withDefaults()

	readerConfig := kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupId,
		Topic:          cfg.StepResultTopic,
		MinBytes:       1,
		MaxBytes:       cfg.MaxBytes,
		MaxWait:        cfg.ReadTimeout,
		CommitInterval: cfg.CommitInterval,
		StartOffset:    cfg.StartOffset,
		// TODO: нужно ли включить?
		ReadLagInterval: -1,
	}

	if cfg.ClientId != "" {
		readerConfig.Dialer = &kafka.Dialer{
			Timeout:   cfg.DialTimeout,
			ClientID:  cfg.ClientId,
			DualStack: true,
		}
	} else {
		readerConfig.Dialer = &kafka.Dialer{
			Timeout:   cfg.DialTimeout,
			DualStack: true,
		}
	}

	return &KafkaStepResultReader{
		reader: kafka.NewReader(readerConfig),
	}, nil
}

func (r *KafkaStepResultReader) Start(ctx context.Context, handler StepResultHandler) error {
	if handler == nil {
		return errors.New("unexpected nil step result handler")
	}

	for {
		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("fetch kafka message: %w", err)
		}

		var event *SagaStepResultEvent
		if err = json.Unmarshal(msg.Value, &event); err != nil {
			// TODO: подумать над отправкой в DLQ
			if commitErr := r.reader.CommitMessages(ctx, msg); commitErr != nil {
				return fmt.Errorf("unmarshal kafka message: %v; commit poisoned message: %w", err, commitErr)
			}
			continue
		}
		if err = handler(ctx, event); err != nil {
			// не коммитим offset, подождем ретрая
			return fmt.Errorf("handle saga.step.result event: %w", err)
		}

		if err = r.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit kafka message: %w", err)
		}
	}
}

func (r *KafkaStepResultReader) Stop() error {
	if r == nil || r.reader == nil {
		return nil
	}
	return r.reader.Close()
}
