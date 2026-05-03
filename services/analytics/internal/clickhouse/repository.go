package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/timickb/sagaflow/services/analytics/internal/domain"
)

type Config struct {
	Addresses []string
	Database  string
	Username  string
	Password  string
}

type Repository struct {
	ch clickhouse.Conn
}

func NewRepository(cfg *Config) (*Repository, error) {
	ch, err := clickhouse.Open(&clickhouse.Options{
		Addr: cfg.Addresses,
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("clickhouse open: %w", err)
	}

	return &Repository{ch: ch}, nil
}

// InsertOrderEvent вставляет запись в таблицу fct_orders
func (r *Repository) InsertOrderEvent(ctx context.Context, event *domain.OrderEvent) error {
	now := time.Now().UTC()

	fctOrder := domain.FctOrder{
		OrderID:     event.OrderId,
		UserID:      event.UserId,
		Status:      event.Status,
		TotalAmount: event.TotalAmount,
		Currnecy:    event.Currency,
		Version:     event.Version,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
		LoadedAt:    now,
	}

	query := `
		INSERT INTO fct_orders (order_id, user_id, status, total_amount, currency, version, created_at, updated_at, loaded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err := r.ch.Exec(ctx, query,
		fctOrder.OrderID,
		fctOrder.UserID,
		fctOrder.Status,
		fctOrder.TotalAmount,
		fctOrder.Currnecy,
		fctOrder.Version,
		fctOrder.CreatedAt,
		fctOrder.UpdatedAt,
		fctOrder.LoadedAt,
	)
	if err != nil {
		return fmt.Errorf("clickhouse insert: %w", err)
	}

	return nil
}

func (r *Repository) Close() error {
	if r.ch != nil {
		return r.ch.Close()
	}
	return nil
}
