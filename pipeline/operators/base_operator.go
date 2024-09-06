package operators

import (
	"context"
	"github.com/pixie-sh/database-helpers-go/database"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/errors-go/utils"
	"reflect"
)

// define the operators types
// this is used to keep the operators compatible for the whole pipeline
const (
	// DatabaseOperatorType expect that the result is a *database.DB
	// any other format shall be implemented in other operator type.
	DatabaseOperatorType uint = 0
)

// QueryParams to be used with all DatabaseOperator
type QueryParams = map[string][]string

// Operator something amazing... or not.
type Operator interface {
	Handle(ctx context.Context, result Result) (Result, error)
	Predicate(ctx context.Context, ignoreOverride bool) bool
	GetType() uint
}

// DatabaseOperator simple struct to hold common DatabaseOperator functions
type DatabaseOperator struct {
	predicateOverride func() bool
	queryParams       QueryParams
	requestParamName  string
}

// GetType return type DatabaseOperatorType
func (b *DatabaseOperator) GetType() uint {
	return DatabaseOperatorType
}

// Predicate basic predicate
func (b *DatabaseOperator) Predicate(_ context.Context, ignoreOverride bool) bool {
	if b.predicateOverride != nil && !ignoreOverride {
		return b.predicateOverride()
	}

	return len(b.queryParams[b.requestParamName]) != 0
}

// getPassable return known passable for DatabaseOperator's
func (b *DatabaseOperator) getPassable(res Result) (*database.DB, error) {
	casted, ok := res.GetPassable().(*database.DB)
	if !ok {
		return nil, errors.New("invalid result passable %s", reflect.TypeOf(res.GetPassable()).String())
	}

	return casted, nil
}

func (b *DatabaseOperator) apply(result Result, tx *database.DB, clause string, args ...interface{}) *database.DB {
	dbRes, ok := result.(*BaseResult)
	if !ok {
		return tx
	}

	if utils.Nil(dbRes.previous) {
		return tx.Where(clause, args...)
	}

	aggregator, ok := dbRes.previous.(*AggregatorOperator)
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
