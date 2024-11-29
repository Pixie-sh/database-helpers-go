package database

import (
	"context"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/env"
	"github.com/pixie-sh/logger-go/logger"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gLog "gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

// DB alias to gorm.DB ptr
type DB = gorm.DB

// LogLevel alias mapping
const (
	LogLevelSilent = gLog.Silent
	LogLevelError  = gLog.Error
	LogLevelWarn   = gLog.Warn
	LogLevelInfo   = gLog.Info
)

// Session alias to gorm.Session
type Session = gorm.Session

// DeletedAt alias to gorm.DeletedAt
type DeletedAt = gorm.DeletedAt

// Orm exposes basic database access functionality
type Orm struct {
	*DB
	configuration *GormDbConfiguration
}

// NewGormDb returns a new Orm instance, that has Gorm nested
func NewGormDb(_ context.Context, cfg *GormDbConfiguration) (*Orm, error) {
	if cfg == nil {
		return nil, errors.New("gorm cfg is nil").WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	var dialect gorm.Dialector
	var replicas []gorm.Dialector
	switch cfg.Driver {
	case MysqlDriver:
		dialect = mysql.Open(cfg.Dsn + "?parseTime=true")
		for _, dsn := range cfg.Replicas.ReplicaDsns {
			replicas = append(replicas, mysql.Open(dsn))
		}
	case PsqlDriver:
		dialect = postgres.Open(cfg.Dsn)
		for _, dsn := range cfg.Replicas.ReplicaDsns {
			replicas = append(replicas, postgres.Open(dsn))
		}
	}

	db, err := gorm.Open(dialect, &gorm.Config{})
	if err != nil {
		return nil, errors.New("unable to open db connection; %s", err.Error()).WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	if len(replicas) > 0 {
		var policy dbresolver.Policy
		switch cfg.Replicas.Policy {
		case "custom":
			policy = cfg.CustomResolverPolicy
			break
		case "random":
		default:
			policy = dbresolver.RandomPolicy{}
		}
		err = db.Use(dbresolver.Register(dbresolver.Config{
			Replicas:          replicas,
			Policy:            policy,
			TraceResolverMode: cfg.Replicas.TraceResolverMode, // print sources/replicas mode in logger
		}))

		if err != nil {
			return nil, errors.New("unable to set replicas; %s", err.Error()).WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
		}
	}

	orm := &Orm{
		DB:            db,
		configuration: cfg,
	}

	if env.IsDebugActive() {
		orm.WithDebug()
	}

	return orm, nil
}

// WithDebug meant to be used for debug purpose only
func (o *Orm) WithDebug() *Orm {
	o.DB = o.Session(&gorm.Session{Logger: log{plog: logger.Logger}}).Debug()
	return o
}

func (o *Orm) Ping() error {
	return o.Exec("SELECT 1").Error
}

func (o *Orm) Close() error {
	//gorm handles it, nothing to do for now
	return nil
}
