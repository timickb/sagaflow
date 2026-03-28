package db

import (
	"errors"
	"fmt"
)

// PostgresConfig Конфигурация соединения с PostgreSQL.
type PostgresConfig struct {
	Host               string            `json:"host" yaml:"host"`
	Name               string            `json:"name" yaml:"name"`
	User               string            `json:"user" yaml:"user"`
	Password           string            `json:"password" yaml:"password"`
	SSLMode            string            `json:"ssl_mode" yaml:"ssl_mode"`
	Port               int               `json:"port" yaml:"port"`
	MaxOpenConnections int               `json:"max_open_connections" yaml:"max_open_connections"`
	MaxIdleConnections int               `json:"max_idle_connections" yaml:"max_idle_connections"`
	AutoMigrate        bool              `json:"auto_migrate" yaml:"auto_migrate"`
	Secondaries        []*PostgresConfig `json:"secondaries" yaml:"secondaries"`
}

// DSNString Сформировать строку для подключения к БД.
func (c *PostgresConfig) DSNString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// Validate Валидировать конфигурацию.
func (c *PostgresConfig) Validate() error {
	if c.Host == "" {
		return errors.New("empty database host")
	}
	if c.Name == "" {
		return errors.New("empty database name")
	}
	if c.Port <= 0 {
		return errors.New("invalid database port")
	}
	return nil
}
