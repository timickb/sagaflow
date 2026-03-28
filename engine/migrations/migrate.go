package migrations

import (
	"embed"

	"github.com/timickb/sagaflow/engine/pkg/db"
)

//go:embed *.sql
var migrations embed.FS

var Migrator = db.NewMigrator(".", migrations)
