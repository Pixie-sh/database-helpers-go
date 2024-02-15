package database

import (
	"context"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/mapper"
)

type FactoryConfiguration struct {
	Mapping map[string]FactoryCreateFn
}

// DefaultFactoryConfiguration default factory configuration that creates tje json logger
var DefaultFactoryConfiguration = FactoryConfiguration{
	Mapping: map[string]FactoryCreateFn{
		GormDriver: createGorm,
	},
}

func createGorm(ctx context.Context, cfg *Configuration) (*Orm, error) {
	var gormCfg GormDbConfiguration
	err := mapper.ObjectToStruct(cfg.Values, &gormCfg)
	if err != nil {
		return nil, errors.New("error mapping struct: %s", err.Error()).WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	instance, err := NewGormDb(ctx, &gormCfg)
	if err != nil {
		return nil, errors.New("error initializing gorm: %s", err.Error()).WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	return instance, nil
}

// Configuration generic
type Configuration struct {
	Driver string      `toml:"driver" ,json:"driver" ,mapstructure:"driver"`
	Values interface{} `toml:"values" ,json:"values" ,mapstructure:"values"`
}

// GormDbConfiguration gorm specific configuration
type GormDbConfiguration struct {
	Driver string `toml:"driver" ,json:"driver" ,mapstructure:"driver"`
	Dsn    string `toml:"dsn" ,json:"dsn" ,mapstructure:"dsn"` //https://github.com/go-sql-driver/mysql#dsn-data-source-name
}
