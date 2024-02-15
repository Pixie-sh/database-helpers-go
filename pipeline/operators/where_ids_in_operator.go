package operators

import (
	"github.com/pixie-sh/database-helpers-go/pipeline"
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
func (op *WhereIdsInOperator) Handle(genericResult pipeline.Result) (pipeline.Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	tx = tx.Where(op.property+" IN (?)", op.queryParams[op.requestParamName])

	genericResult.WithPassable(tx)
	return genericResult, tx.Error
}
