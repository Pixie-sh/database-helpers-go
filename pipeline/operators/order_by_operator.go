package operators

import (
	"github.com/pixie-sh/database-helpers-go/pipeline"
	"github.com/pixie-sh/errors-go"
	"strings"
)

// OrderByOperator something amazing... or not.
type OrderByOperator struct {
	DatabaseOperator

	defaultSortProperties []string
	sortableProperties    []string
	acceptsRequestInput   bool
}

// NewOrderByOperator something amazing, is it?
func NewOrderByOperator(queryParams QueryParams, acceptsRequestInput bool, defaultSortProperty []string, sortableProperties ...string) *OrderByOperator {
	newOperator := &OrderByOperator{}
	newOperator.defaultSortProperties = defaultSortProperty
	newOperator.sortableProperties = sortableProperties
	newOperator.acceptsRequestInput = acceptsRequestInput
	newOperator.queryParams = queryParams
	newOperator.requestParamName = "sort_by"
	newOperator.predicateOverride = newOperator.predicate
	return newOperator
}

// Predicate something amazing... who knows....
func (op *OrderByOperator) predicate() bool {
	return op.acceptsRequestInput == false || len(op.queryParams[op.requestParamName]) != 0
}

// getAllSortConditions something amazing... who knows....
func (op *OrderByOperator) getAllSortConditions(params QueryParams, requestParam string) []string {
	conditions := op.defaultSortProperties
	resultingConditions := make([]string, 0)

	conditions = params[requestParam]

	for _, condition := range conditions {
		trimmed := strings.TrimSpace(condition)
		if trimmed[0] == '-' || trimmed[0] == '+' {
			if contains(op.sortableProperties, trimLeftChar(trimmed)) {
				resultingConditions = append(resultingConditions, trimmed)
			}
		} else if contains(op.sortableProperties, trimmed) {
			resultingConditions = append(resultingConditions, "+"+trimmed)
		}
	}

	return resultingConditions
}

// Handle something amazing... who knows....
func (op *OrderByOperator) Handle(genericResult pipeline.Result) (pipeline.Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}
	sortingTerms := op.getAllSortConditions(op.queryParams, op.requestParamName)

	for _, term := range sortingTerms {
		order := " asc"
		if term[0] == '-' {
			order = " desc"
		}

		tx = tx.Order(trimLeftChar(term) + order)
	}

	genericResult.WithPassable(tx)
	return genericResult, tx.Error
}

func trimLeftChar(s string) string {
	for i := range s {
		if i > 0 {
			return s[i:]
		}
	}
	return s[:0]
}
