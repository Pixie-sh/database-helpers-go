// Package sql holds the GORM/SQL-specific pipeline operators. Every operator
// in this package expects the pipeline passable to be a *database.DB.
package sql

import (
	"context"
	"reflect"

	"github.com/pixie-sh/database-helpers-go/database"
	databaserrors "github.com/pixie-sh/database-helpers-go/errors"
	"github.com/pixie-sh/database-helpers-go/pipeline/operators/core"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/errors-go/utils"
)

// Convenience aliases so call sites inside this package can keep using the
// short historical names without spelling out core.* on every line.
type (
	Operator    = core.Operator
	Result      = core.Result
	BaseResult  = core.BaseResult
	QueryParams = core.QueryParams
)

// DatabaseOperator is the shared base struct embedded by every SQL operator.
// It carries the request-scoped query params, an optional name pointing at
// the relevant param, and an optional predicate override.
type DatabaseOperator struct {
	predicateOverride func() bool
	queryParams       QueryParams
	requestParamName  string
}

// GetType returns core.DatabaseOperatorType, satisfying the Operator contract.
func (b *DatabaseOperator) GetType() uint {
	return core.DatabaseOperatorType
}

// Predicate reports whether the operator should run for the current request.
// When no override is configured the default behaviour is "run only when the
// relevant query param is present and non-empty".
func (b *DatabaseOperator) Predicate(_ context.Context, ignoreOverride bool) bool {
	if b.predicateOverride != nil && !ignoreOverride {
		return b.predicateOverride()
	}

	return len(b.queryParams[b.requestParamName]) != 0
}

// getPassable extracts a *database.DB from a Result, erroring out if the
// passable carries something else.
func (b *DatabaseOperator) getPassable(res Result) (*database.DB, error) {
	casted, ok := res.GetPassable().(*database.DB)
	if !ok {
		return nil, errors.New("invalid result passable %s", reflect.TypeOf(res.GetPassable()).String()).WithErrorCode(databaserrors.InvalidResultPassableErrorCode)
	}

	return casted, nil
}

// apply appends a WHERE clause to tx, honouring the boolean combinator hint
// left by a prior AggregatorOperator on the result.
func (b *DatabaseOperator) apply(result Result, tx *database.DB, clause string, args ...interface{}) *database.DB {
	dbRes, ok := result.(*BaseResult)
	if !ok {
		return tx
	}

	previous := dbRes.Previous()
	if utils.Nil(previous) {
		return tx.Where(clause, args...)
	}

	aggregator, ok := previous.(*AggregatorOperator)
	if !ok {
		return tx.Where(clause, args...)
	}

	switch aggregator.aggregator {
	case AggregatorConditionOR:
		return tx.Or(clause, args...)
	default:
		return tx.Where(clause, args...)
	}
}
