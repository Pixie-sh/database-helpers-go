package database

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/mapper"
	"gorm.io/plugin/dbresolver"
)

type FactoryConfiguration struct {
	Mapping map[string]FactoryCreateFn[*any]
}

// DefaultFactoryConfiguration default factory configuration that creates tje json logger
var DefaultFactoryConfiguration = FactoryConfiguration{
	Mapping: map[string]FactoryCreateFn[*any]{
		GormDriver: createGorm,
		PgxDriver:  createGorm,
	},
}

func createGorm(ctx context.Context, cfg *Configuration) (*Orm, error) {
	var gormCfg GormDbConfiguration
	err := mapper.ObjectToStruct(cfg.Values, &gormCfg)
	if err != nil {
		return nil, errors.New("error mapping struct: %s", err.Error()).WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	switch gormCfg.Replicas.Policy {
	case "custom":
		gormCfg.CustomResolverPolicy = cfg.Custom["custom_resolver"].(dbresolver.Policy)
		break
	default:
		gormCfg.Replicas.Policy = "random"
	}

	instance, err := NewGormDb(ctx, &gormCfg)
	if err != nil {
		return nil, errors.New("error initializing gorm: %s", err.Error()).WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	return instance, nil
}

func createPGX(ctx context.Context, cfg *Configuration) (*pgxpool.Pool, error) {
	var gormCfg GormDbConfiguration
	err := mapper.ObjectToStruct(cfg.Values, &gormCfg)
	if err != nil {
		return nil, errors.New("error mapping struct: %s", err.Error()).WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	switch gormCfg.Replicas.Policy {
	case "custom":
		gormCfg.CustomResolverPolicy = cfg.Custom["custom_resolver"].(dbresolver.Policy)
		break
	default:
		gormCfg.Replicas.Policy = "random"
	}

	instance, err := NewGormDb(ctx, &gormCfg)
	if err != nil {
		return nil, errors.New("error initializing gorm: %s", err.Error()).WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	return instance, nil
}

// Configuration generic
type Configuration struct {
	Driver string                 `toml:"driver" json:"driver" mapstructure:"driver"`
	Values interface{}            `toml:"values" json:"values" mapstructure:"values"`
	Custom map[string]interface{} `toml:"-" json:"-" mapstructure:"-"` //to hold injectable struct into gorm
}

// GormDbConfiguration gorm specific configuration
type GormDbConfiguration struct {
	Driver               string                      `toml:"driver" json:"driver" mapstructure:"driver"`
	Dsn                  string                      `toml:"dsn" json:"dsn" mapstructure:"dsn"`                //https://github.com/go-sql-driver/mysql#dsn-data-source-name
	Replicas             GormReplicasDbConfiguration `toml:"replicas" json:"replicas" mapstructure:"replicas"` //https://gorm.io/docs/dbresolver.html
	CustomResolverPolicy dbresolver.Policy           `json:"-" mapstructure:"custom_resolver_policy"`
}

type GormReplicasDbConfiguration struct {
	ReplicaDsns       []string `toml:"replica_dsns" json:"replica_dsns" mapstructure:"replica_dsns"`
	TraceResolverMode bool     `toml:"trace_resolver" json:"trace_resolver" mapstructure:"trace_resolver"`
	Policy            string   `toml:"policy" json:"policy" mapstructure:"policy"`
}
