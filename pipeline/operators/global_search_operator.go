package operators

import (
	"github.com/pixie-sh/database-helpers-go/pipeline"
	"github.com/pixie-sh/errors-go"
	"strings"
)

// GlobalSearchOperator something amazing... or not.
type GlobalSearchOperator struct {
	DatabaseOperator

	searchableProperties []string
}

// NewGlobalSearchOperator something amazing, is it?
func NewGlobalSearchOperator(queryParams QueryParams, requestParamName string, searchableProperties ...string) *GlobalSearchOperator {
	newOperator := new(GlobalSearchOperator)
	newOperator.queryParams = queryParams
	newOperator.searchableProperties = searchableProperties
	newOperator.requestParamName = requestParamName
	return newOperator
}

// Handle something amazing... who knows....
func (op *GlobalSearchOperator) Handle(genericResult pipeline.Result) (pipeline.Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	searchTerm := "%" + strings.Join(op.queryParams[op.requestParamName], ",") + "%"

	var fields []string
	var values []interface{}

	for _, prop := range op.searchableProperties {
		fields = append(fields, "`"+prop+"` LIKE ?")
		values = append(values, searchTerm)
	}

	genericResult.WithPassable(tx.Where(strings.Join(fields, " OR "), values...))
	return genericResult, tx.Error
}
