package operators

import (
	"context"
	"github.com/pixie-sh/errors-go"
)

// WhereIdsInOperator something amazing... or not.
type WhereIdsInOperator struct {
	DatabaseOperator

	property string
}

// NewWhereIdsInOperator something amazing, is it?
func NewWhereIdsInOperator(queryParams QueryParams, property string, requestParamName string) *WhereIdsInOperator {
	newOperator := &WhereIdsInOperator{}
	newOperator.queryParams = queryParams
	newOperator.property = property
	newOperator.requestParamName = requestParamName
	return newOperator
}

// Handle something amazing... who knows....
func (op *WhereIdsInOperator) Handle(ctx context.Context, genericResult Result) (Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	tx = tx.Where(op.property+" IN (?)", op.queryParams[op.requestParamName])

	genericResult.WithPassable(tx)
	return genericResult, tx.Error
}
