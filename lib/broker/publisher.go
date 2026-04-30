package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaStepResultWriter - publisher для записи в топик saga.step.result
type KafkaStepResultWriter struct {
	writer *kafka.Writer
	topic  string
}

// NewKafkaStepResultWriter - создать writer для записи в топик результатов шагов саги
func NewKafkaStepResultWriter(cfg *KafkaConfig) (*KafkaStepResultWriter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate kafka config: %w", err)
	}
	cfg = cfg.withDefaults()

	dialer := &net.Dialer{Timeout: cfg.DialTimeout}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.StepResultTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		Async:        false,
		Transport: &kafka.Transport{
			Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, addr)
			},
		},
	}

	return &KafkaStepResultWriter{
		writer: writer,
		topic:  cfg.StepResultTopic,
	}, nil
}

// Publish - публикует событие в топик saga.step.result
func (w *KafkaStepResultWriter) Publish(ctx context.Context, event *SagaStepResultEvent) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	if event.Ref.SagaId.String() == "" {
		return fmt.Errorf("event saga_id is required")
	}
	if event.Ref.StepName == "" {
		return fmt.Errorf("event step_name is required")
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("%s-%s", event.Ref.SagaId.String(), event.Ref.StepName)),
		Value: data,
	}

	if err = w.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write kafka message: %w", err)
	}

	return nil
}

// PublishBatch - публикует несколько событий в топик
func (w *KafkaStepResultWriter) PublishBatch(ctx context.Context, events []*SagaStepResultEvent) error {
	if len(events) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(events))

	for _, event := range events {
		if event == nil {
			continue
		}
		if event.Ref.SagaId.String() == "" {
			continue
		}
		if event.Ref.StepName == "" {
			continue
		}

		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("marshal event: %w", err)
		}

		messages = append(messages, kafka.Message{
			Key:   []byte(fmt.Sprintf("%s-%s", event.Ref.SagaId.String(), event.Ref.StepName)),
			Value: data,
		})
	}

	if len(messages) == 0 {
		return nil
	}

	if err := w.writer.WriteMessages(ctx, messages...); err != nil {
		return fmt.Errorf("write kafka messages: %w", err)
	}

	return nil
}

// Close - закрывает writer
func (w *KafkaStepResultWriter) Close() error {
	if w == nil || w.writer == nil {
		return nil
	}
	return w.writer.Close()
}

// StepResultWriter - интерфейс для публикации событий результатов шагов
type StepResultWriter interface {
	Publish(ctx context.Context, event *SagaStepResultEvent) error
	PublishBatch(ctx context.Context, events []*SagaStepResultEvent) error
	Close() error
}

// SafeStepResultWriter - thread-safe обертка над StepResultWriter
type SafeStepResultWriter struct {
	writer StepResultWriter
	mu     sync.Mutex
}

// NewSafeStepResultWriter - создает потокобезопасную обертку
func NewSafeStepResultWriter(writer StepResultWriter) *SafeStepResultWriter {
	return &SafeStepResultWriter{
		writer: writer,
	}
}

// Publish - потокобезопасная публикация события
func (w *SafeStepResultWriter) Publish(ctx context.Context, event *SagaStepResultEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Publish(ctx, event)
}

// PublishBatch - потокобезопасная публикация батча событий
func (w *SafeStepResultWriter) PublishBatch(ctx context.Context, events []*SagaStepResultEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.PublishBatch(ctx, events)
}

// Close - закрывает underlying writer
func (w *SafeStepResultWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Close()
}
