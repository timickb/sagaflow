package db

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type transactor struct {
	db *Database
}

// NewTransactor Создать сущность для запуска gorm транзакций.
func NewTransactor(db *Database) *transactor {
	return &transactor{db: db}
}

// Transaction Начать gorm транзакцию.
func (t *transactor) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	db := fetchDbFromCtx(ctx)
	if db != nil {
		return errors.New("transaction is already started")
	}

	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(putDbToCtx(ctx, &Database{tx}))
	})
}
