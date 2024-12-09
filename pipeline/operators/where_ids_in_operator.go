package operators

import (
	"context"
	"github.com/pixie-sh/errors-go"
	"strings"
)

// WhereIdsInOperator something amazing... or not.
type WhereIdsInOperator struct {
	DatabaseOperator

	property       string
	maxNumberOfIds int
}

// NewWhereIdsInOperator something amazing, is it?
func NewWhereIdsInOperator(queryParams QueryParams, property string, requestParamName string, maxNumberOfIds ...int) *WhereIdsInOperator {
	newOperator := &WhereIdsInOperator{}
	newOperator.queryParams = queryParams
	newOperator.property = property
	newOperator.requestParamName = requestParamName
	if len(maxNumberOfIds) > 0 {
		newOperator.maxNumberOfIds = maxNumberOfIds[0]
	} else {
		newOperator.maxNumberOfIds = -1
	}

	if newOperator.queryParams == nil {
		newOperator.queryParams = make(QueryParams)
	}

	if len(newOperator.queryParams[newOperator.requestParamName]) == 0 {
		newOperator.queryParams[newOperator.requestParamName] = []string{""}
	}

	return newOperator
}

// Handle something amazing... who knows....
func (op *WhereIdsInOperator) Handle(ctx context.Context, genericResult Result) (Result, error) {
	ids := strings.Split(op.queryParams[op.requestParamName][0], ",")
	if op.maxNumberOfIds > 0 && len(ids) > op.maxNumberOfIds {
		return nil, errors.New("number of ids exceeds the maximum allowed")
	}
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	tx = tx.Where(op.property+" IN (?)", ids)

	genericResult.WithPassable(tx)
	return genericResult, tx.Error
}
