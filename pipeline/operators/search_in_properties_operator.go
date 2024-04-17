package operators

import (
	"github.com/pixie-sh/database-helpers-go/pipeline"
	"github.com/pixie-sh/errors-go"
	"strings"
)

// SearchInPropertiesOperator something amazing... or not.
type SearchInPropertiesOperator struct {
	DatabaseOperator

	properties []string
}

// NewSearchInPropertiesOperator something amazing, is it?
func NewSearchInPropertiesOperator(queryParams QueryParams, requestParamName string, properties ...string) *SearchInPropertiesOperator {
	newOperator := &SearchInPropertiesOperator{}
	newOperator.properties = properties
	newOperator.requestParamName = requestParamName
	newOperator.queryParams = queryParams
	return newOperator
}

// getAllValidConditions something amazing... who knows....
func (op *SearchInPropertiesOperator) getAllValidConditions(params QueryParams) []string {
	query := strings.Split(strings.Join(params[op.requestParamName], ","), " AND ")
	resultingConditions := make([]string, 0)

	for _, condition := range query {
		conditionSplit := strings.SplitN(condition, ":", 2)

		if contains(op.properties, conditionSplit[0]) {
			resultingConditions = append(resultingConditions, condition)
		}
	}

	return resultingConditions
}

// Handle something amazing... who knows....
func (op *SearchInPropertiesOperator) Handle(genericResult pipeline.Result) (pipeline.Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	searchTerms := op.getAllValidConditions(op.queryParams)

	for _, term := range searchTerms {
		termSplit := strings.SplitN(term, ":", 2)
		tx = tx.Where("LOWER("+termSplit[0]+") LIKE ?", "%"+termSplit[1]+"%")
	}

	genericResult.WithPassable(tx)
	return genericResult, tx.Error
}

// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
