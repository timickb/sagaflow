package migrations

import (
	"embed"

	"github.com/timickb/sagaflow/lib/db"
)

//go:embed *.sql
var migrations embed.FS

var Migrator = db.NewMigrator(".", migrations)
