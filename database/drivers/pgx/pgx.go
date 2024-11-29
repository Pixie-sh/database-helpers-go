package pgx

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/env"
	"github.com/pixie-sh/logger-go/logger"
)

// PgxDB alias to pgxpool.Pool ptr
type PgxDB = pgxpool.Pool

// Pgx exposes basic database access functionality
type Pgx struct {
	*PgxDB
	configuration *pgxpool.Config
	ctx           context.Context
}

// NewPGXPool returns a new PgxPool instance, that has Gorm nested
func NewPGXPool(ctx context.Context, cfg *pgxpool.Config) (*Pgx, error) {
	if cfg == nil {
		return nil, errors.New("pgx cfg is nil").WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.New("unable to open db connection; %s", err.Error()).WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	orm := &Pgx{
		PgxDB:         db,
		configuration: cfg,
	}

	if env.IsDebugActive() {
		orm.withDebug()
	}

	return orm, nil
}

// WithDebug meant to be used for debug purpose only
func WithDebug(db *PgxDB) *PgxDB {
	// return db.Session(&gorm.Session{Logger: log{plog: logger.Logger}}).Debug()
	logger.Logger.Warn("Pgx doesn't support debug mode like this. Please use DB configs for it.")
	return db
}

// withDebug meant to be used for debug purpose only
func (o *Pgx) withDebug() *Pgx {
	o.PgxDB = WithDebug(o.PgxDB)
	return o
}

func (o *Pgx) Ping() error {
	_, err := o.Exec(o.ctx, "SELECT 1")
	return err
}

func (o *Pgx) Close() error {
	//gorm handles it, nothing to do for now
	return nil
}
