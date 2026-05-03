package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

const (
	OrdersEventsTopic = "analytics.orders.events"
)

type OrderHandler interface {
	HandleOrderEvent(ctx context.Context, event []byte) error
}

type KafkaConsumer struct {
	reader *kafka.Reader
}

type KafkaConsumerConfig struct {
	Brokers []string
	GroupID string
	Topic   string
}

func NewKafkaConsumer(cfg *KafkaConsumerConfig) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    cfg.Topic,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	return &KafkaConsumer{reader: reader}
}

// Start запускает консьюмер в бесконечном цикле
func (c *KafkaConsumer) Start(ctx context.Context, handler OrderHandler) error {
	log.Info().Str("topic", OrdersEventsTopic).Msg("Starting kafka consumer for order events")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Consumer context cancelled, stopping")
			return nil
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				log.Error().Err(err).Msg("Failed to fetch message")
				continue
			}

			// Парсим сообщение как событие заказа
			var orderData map[string]interface{}
			if err := json.Unmarshal(msg.Value, &orderData); err != nil {
				log.Error().
					Err(err).
					Str("value", string(msg.Value)).
					Msg("Failed to unmarshal order event")

				// Коммитим даже битое сообщение, чтобы не застревать
				if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
					log.Error().Err(commitErr).Msg("Failed to commit poisoned message")
				}
				continue
			}

			// Передаем raw bytes в handler для обработки
			if err := handler.HandleOrderEvent(ctx, msg.Value); err != nil {
				log.Error().
					Err(err).
					Str("order_id", fmt.Sprintf("%v", orderData["order_id"])).
					Msg("Failed to handle order event")
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Error().Err(err).Msg("Failed to commit message")
			}
		}
	}
}

func (c *KafkaConsumer) Stop() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
