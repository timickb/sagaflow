package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

const (
	fetchFailedBackoff = time.Second
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
		Brokers:         cfg.Brokers,
		GroupID:         cfg.GroupId,
		Topic:           cfg.StepResultTopic,
		MinBytes:        1,
		MaxBytes:        cfg.MaxBytes,
		MaxWait:         cfg.ReadTimeout,
		CommitInterval:  cfg.CommitInterval,
		StartOffset:     cfg.StartOffset,
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

	log.Info().Msg("Started kafka consumer")

	for {
		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Info().Msg("Consumer received context done")
				return nil
			}

			log.Error().Err(err).Msg("Failed to fetch kafka message")
			time.Sleep(fetchFailedBackoff)
			continue
		}

		var event SagaStepResultEvent
		if err = json.Unmarshal(msg.Value, &event); err != nil {
			log.Error().
				Err(err).
				Int("partition", msg.Partition).
				Int64("offset", msg.Offset).
				Msg("Failed to unmarshal kafka message")

			if commitErr := r.reader.CommitMessages(ctx, msg); commitErr != nil {
				log.Error().Err(commitErr).Msg("Failed to commit poisoned kafka message")
			}
			continue
		}
		if err = handler(ctx, &event); err != nil {
			log.Error().
				Err(err).
				Int("partition", msg.Partition).
				Int64("offset", msg.Offset).
				Msg("Failed to handle step result event")
			continue
		}

		if err = r.reader.CommitMessages(ctx, msg); err != nil {
			log.Error().
				Err(err).
				Int("partition", msg.Partition).
				Int64("offset", msg.Offset).
				Msg("Failed to commit handled kafka message")
		}
	}

}

func (r *KafkaStepResultReader) Stop() error {
	if r == nil || r.reader == nil {
		return nil
	}
	return r.reader.Close()
}
