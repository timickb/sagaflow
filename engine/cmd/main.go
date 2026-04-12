package main

import (
	"flag"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/timickb/sagaflow/engine/internal/config"
	"github.com/timickb/sagaflow/engine/internal/infra"
	"github.com/timickb/sagaflow/lib/utils"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	configPath := flag.String("cfg", "config.yaml", "config file")
	flag.Parse()

	if utils.IsStrNilOrEmpty(configPath) {
		log.Fatal().Msg("Config file path is required")
	}

	cfg, err := config.NewFromFile(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config file")
	}

	builder, err := infra.NewBuilder(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build dependencies")
	}

	if err = builder.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start engine")
	}
}
