package operators

import (
	"fmt"
	"github.com/pixie-sh/database-helpers-go/pipeline"
	"github.com/pixie-sh/errors-go"
	"strings"
)

// JsonSearchOperator something amazing... or not.
type JsonSearchOperator struct {
	DatabaseOperator

	whereClause string
	whereFormat string
}

// NewJsonSearchOperator something amazing, is it?
func NewJsonSearchOperator(queryParams QueryParams, requestParamName string, whereClause string, whereFormat string) *JsonSearchOperator {
	newOperator := new(JsonSearchOperator)
	newOperator.queryParams = queryParams
	newOperator.whereClause = whereClause
	newOperator.whereFormat = whereFormat
	newOperator.requestParamName = requestParamName
	return newOperator
}

// Handle something amazing... who knows....
func (op *JsonSearchOperator) Handle(genericResult pipeline.Result) (pipeline.Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	genericResult.WithPassable(tx.Where(op.whereClause, fmt.Sprintf(op.whereFormat, strings.Join(op.queryParams[op.requestParamName], ","))))

	return genericResult, tx.Error
}
