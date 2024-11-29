package database

import (
	"context"
	"fmt"
	"github.com/pixie-sh/errors-go"
	"reflect"
)

type FactoryCreateFn = func(ctx context.Context, configuration *Configuration) (any, error)

type Factory struct {
	Mapping map[string]FactoryCreateFn
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
func (f Factory) Create(ctx context.Context, configuration *Configuration) (any, error) {
	fn, exist := f.Mapping[configuration.Driver]
	if !exist {
		return nil, errors.New(
			"unknown database driver %s; unable to create orm",
			configuration.Driver).WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	return fn(ctx, configuration)
}

func Create[T interface{}](ctx context.Context, configuration *Configuration, withCustomFactory ...Factory) (T, error) {
	var t T
	var ok bool

	var factory = FactoryInstance
	if len(withCustomFactory) > 0 {
		factory = withCustomFactory[0]
	}

	a, err := factory.Create(ctx, configuration)
	if err != nil {
		return t, err
	}

	t, ok = a.(T)
	if !ok {
		return t, errors.New("unable to create orm from type %s", reflect.TypeOf(a).String())
	}

	return t, nil
}
