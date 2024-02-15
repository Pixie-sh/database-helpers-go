package database

import (
	"context"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/pixie-sh/errors-go"
)

// MigrationDefaults alias to gormigrate default options
var MigrationDefaults = gormigrate.DefaultOptions

// MigratorConfiguration alias to gormigrate options
type MigratorConfiguration = gormigrate.Options

// Migration migration alias
type Migration = gormigrate.Migration

// Migrator executor for migrations
type Migrator struct {
	*gormigrate.Gormigrate
}

// NewMigrator returns a new Migrator pointer
func NewMigrator(_ context.Context, configuration *MigratorConfiguration, orm *Orm, migrations ...*Migration) (*Migrator, error) {
	if len(migrations) == 0 {
		return nil, errors.New("migrations are empty unable to create migrator").WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	return &Migrator{gormigrate.New(orm.DB, configuration, migrations)}, nil
}
