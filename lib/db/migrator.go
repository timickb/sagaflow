package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"io/fs"
)

// Migrator Сущность для осуществления миграций.
type Migrator struct {
	path string
	fSys fs.FS
}

func NewMigrator(path string, fSys fs.FS) *Migrator {
	return &Migrator{
		path: path,
		fSys: fSys,
	}
}

// Migrate Применить все миграции к БД.
func (m *Migrator) Migrate(db *sql.DB, dbName string) error {
	p, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: "_migrations",
	})
	if err != nil {
		return fmt.Errorf("create postgres instance: %w", err)
	}

	data, err := iofs.New(m.fSys, m.path)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	mg, err := migrate.NewWithInstance("fs", data, dbName, p)
	if err != nil {
		return fmt.Errorf("create migrator instance: %w", err)
	}

	if err = mg.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("migrator up: %w", err)
	}

	return nil
}
