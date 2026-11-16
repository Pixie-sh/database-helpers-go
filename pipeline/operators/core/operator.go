// Package core holds the foundational, stack-agnostic primitives shared by
// every operator implementation in pipeline/operators/{sql,elastic,...}.
//
// It exists to break the import cycle that would otherwise form between the
// concrete operator subpackages and the historical pipeline/operators facade:
// concrete packages import core for the Operator/Result interfaces;
// pipeline/operators re-exports core symbols (and selectively aliases the
// concrete packages) so existing consumers continue to compile unchanged.
package core

import "context"

// Operator type identifiers, used by Pipeline to ensure every step in the
// pipeline expects a compatible passable type.
const (
	// DatabaseOperatorType expects the passable to be a *database.DB.
	DatabaseOperatorType uint = 0
	// ElasticOperatorType expects the passable to be an Elasticsearch
	// request builder.
	ElasticOperatorType uint = 1
)

// QueryParams is the standard URL-style query-parameter shape used by
// every operator implementation.
type QueryParams = map[string][]string

// Operator is the contract every pipeline operator implements.
type Operator interface {
	Handle(ctx context.Context, result Result) (Result, error)
	Predicate(ctx context.Context, ignoreOverride bool) bool
	GetType() uint
}
