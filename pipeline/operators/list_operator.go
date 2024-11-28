package operators

import (
	"context"
	"github.com/pixie-sh/errors-go"
)

// ListOperator something amazing... or not.
type ListOperator struct {
	DatabaseOperator

	limitRecords int
	dest         interface{}
}

// NewListOperator something amazing, is it?
func NewListOperator(queryParams QueryParams, dest interface{}, limitRecords int) *ListOperator {
	newOperator := &ListOperator{}
	newOperator.dest = dest
	newOperator.limitRecords = limitRecords
	newOperator.queryParams = queryParams
	newOperator.predicateOverride = newOperator.predicate
	return newOperator
}

func (op *ListOperator) predicate() bool {
	return true
}

// Handle something amazing... who knows....
func (op *ListOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	var paginateResult UntypedListedResult

	tx.
		Limit(op.limitRecords).
		Find(op.dest).
		Limit(-1).
		Count(&paginateResult.TotalResults)

	paginateResult.QueryParams = op.queryParams

	genericResult.WithPassable(&paginateResult)
	return genericResult, tx.Error
}
