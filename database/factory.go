package database

import (
	"context"
	"fmt"
	"github.com/pixie-sh/errors-go"
)

type FactoryCreateFn[T] = func(ctx context.Context, configuration *Configuration) (*T, error)

type Factory struct {
	Mapping map[string]FactoryCreateFn[*any]
}

func NewFactory(_ context.Context, config FactoryConfiguration) (Factory, error) {
	if config.Mapping == nil {
		return Factory{}, fmt.Errorf("unable to creater factory, configuration is missing mappings")
	}

	return Factory{
		Mapping: config.Mapping,
	}, nil
}

// Create returns an instance of orm or error if unable to
func (f Factory) Create(ctx context.Context, configuration *Configuration) (*any, error) {
	fn, exist := f.Mapping[configuration.Driver]
	if !exist {
		return nil, errors.New(
			"unknown database driver %s; unable to create orm",
			configuration.Driver).WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	return fn(ctx, configuration)
}
