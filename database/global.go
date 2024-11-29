package database

import (
	"context"
)

// FactoryInstance exported default db factory
var FactoryInstance, _ = NewFactory(context.Background(), DefaultFactoryConfiguration)
