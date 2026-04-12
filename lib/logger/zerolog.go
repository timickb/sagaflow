package logger

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

type Config struct {
	Level   string `yaml:"level"`
	Pretty  bool   `yaml:"pretty"`
	Service string `yaml:"service"`
}

func New(cfg *Config) zerolog.Logger {
	level, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	var out io.Writer = os.Stdout

	if cfg.Pretty {
		out = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05.000",
			NoColor:    false,
		}
	}

	logger := zerolog.New(out).
		With().
		Timestamp().
		Str("service", cfg.Service).
		Logger()

	return logger
}
