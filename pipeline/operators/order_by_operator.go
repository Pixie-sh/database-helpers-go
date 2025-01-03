package operators

import (
	"context"
	"github.com/pixie-sh/database-helpers-go/pipeline/operators/models"
	"github.com/pixie-sh/errors-go"
	"strings"
)

type OrderByOperator struct {
	DatabaseOperator
	defaultSortProperties []string
	sortableProperties    map[string]models.SearchableProperty
	acceptsRequestInput   bool
}

func NewOrderByOperator(queryParams QueryParams, acceptsRequestInput bool, defaultSortProperties []string, sortableProperties ...models.SearchableProperty) *OrderByOperator {
	newOperator := &OrderByOperator{
		defaultSortProperties: defaultSortProperties,
		sortableProperties:    make(map[string]models.SearchableProperty),
		acceptsRequestInput:   acceptsRequestInput,
		DatabaseOperator: DatabaseOperator{
			queryParams:      queryParams,
			requestParamName: "sort_by",
		},
	}

	for _, prop := range sortableProperties {
		newOperator.sortableProperties[prop.Field] = prop
	}

	newOperator.predicateOverride = newOperator.predicate
	return newOperator
}

func (op *OrderByOperator) predicate() bool {
	return true
}

func (op *OrderByOperator) getAllSortConditions(params QueryParams) []string {
	conditions := op.defaultSortProperties
	if op.acceptsRequestInput && len(params[op.requestParamName]) != 0 {
		conditions = params[op.requestParamName]
	}

	var resultingConditions []string
	for _, condition := range conditions {
		trimmed := strings.TrimSpace(condition)
		field := trimmed
		order := "ASC"

		if trimmed[0] == '-' {
			field = trimmed[1:]
			order = "DESC"
		} else if trimmed[0] == '+' {
			field = trimmed[1:]
		}

		if prop, ok := op.sortableProperties[field]; ok {
			resultingConditions = append(resultingConditions, op.buildOrderByClause(prop, order))
		}
	}

	return resultingConditions
}

func (op *OrderByOperator) buildOrderByClause(prop models.SearchableProperty, order string) string {
	switch prop.Type {
	case "date":
		return op.buildDateOrderByClause(prop, order)
	case "text", "varchar":
		return op.buildTextOrderByClause(prop, order)
	case "int", "bigint":
		return op.buildNumericOrderByClause(prop, order)
	case "bool":
		return op.buildBoolOrderByClause(prop, order)
	case "uuid":
		return op.buildUUIDOrderByClause(prop, order)
	case "enum":
		return op.buildEnumOrderByClause(prop, order)
	default:
		return prop.Field + " " + order
	}
}

func (op *OrderByOperator) buildNumericOrderByClause(prop models.SearchableProperty, order string) string {
	return prop.Field + " " + order
}

func (op *OrderByOperator) buildBoolOrderByClause(prop models.SearchableProperty, order string) string {
	return prop.Field + " " + order
}

func (op *OrderByOperator) buildUUIDOrderByClause(prop models.SearchableProperty, order string) string {
	return prop.Field + " " + order
}

func (op *OrderByOperator) buildEnumOrderByClause(prop models.SearchableProperty, order string) string {
	return prop.Field + " " + order
}

func (op *OrderByOperator) buildDateOrderByClause(prop models.SearchableProperty, order string) string {
	if prop.Format != "" {
		return "TO_DATE(" + prop.Field + ", '" + prop.Format + "') " + order
	}
	return prop.Field + " " + order
}

func (op *OrderByOperator) buildTextOrderByClause(prop models.SearchableProperty, order string) string {
	return "LOWER(" + prop.Field + ") " + order
}

func (op *OrderByOperator) Handle(ctx context.Context, genericResult Result) (Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	sortingTerms := op.getAllSortConditions(op.queryParams)

	for _, term := range sortingTerms {
		tx = tx.Order(term)
	}

	genericResult.WithPassable(tx)
	return genericResult, tx.Error
}
