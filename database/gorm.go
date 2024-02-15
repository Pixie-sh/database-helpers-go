package database

import (
	"context"
	"github.com/pixie-sh/errors-go"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gLog "gorm.io/gorm/logger"
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
	switch cfg.Driver {
	case MysqlDriver:
		dialect = mysql.Open(cfg.Dsn + "?parseTime=true")
	case PsqlDriver:
		dialect = postgres.Open(cfg.Dsn)
	}

	db, err := gorm.Open(dialect, &gorm.Config{})
	if err != nil {
		return nil, errors.New(err.Error()).WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	return &Orm{
		DB:            db,
		configuration: cfg,
	}, nil
}

// WithDebug meant to be used for debug purpose only
func WithDebug(db *DB) *DB {
	newLogger := gLog.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		gLog.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  gLog.Info,   // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)

	return db.Session(&gorm.Session{Logger: newLogger}).Debug()
}
