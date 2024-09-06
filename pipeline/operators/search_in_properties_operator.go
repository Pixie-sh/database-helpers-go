package operators

import (
	"context"
	"github.com/google/uuid"
	"github.com/pixie-sh/database-helpers-go/pipeline/operators/models"
	"github.com/pixie-sh/errors-go"
	"strconv"
	"strings"
	"time"
)

type SearchInPropertiesOperator struct {
	DatabaseOperator
	properties map[string]models.SearchableProperty
}

func NewSearchInPropertiesOperator(queryParams QueryParams, requestParamName string, properties ...models.SearchableProperty) *SearchInPropertiesOperator {
	newOperator := &SearchInPropertiesOperator{
		properties: make(map[string]models.SearchableProperty),
	}
	newOperator.requestParamName = requestParamName
	newOperator.queryParams = queryParams

	for _, prop := range properties {
		newOperator.properties[prop.Field] = prop
	}

	return newOperator
}

func (op *SearchInPropertiesOperator) Handle(ctx context.Context, genericResult Result) (Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	searchTerms := op.getAllValidConditions(op.queryParams)

	var conditions []queryCondition

	for _, value := range searchTerms {
		if prop, ok := op.properties[value.Query]; ok {
			condition, parsedValue := op.buildCondition(prop, value.Value)
			if condition != "" {
				conditions = append(conditions, queryCondition{
					Condition:  condition,
					Value:      parsedValue,
					Aggregator: aggregatorFromString(value.Aggregator),
				})
			}
		}
	}

	if len(conditions) > 0 {
		whereClause, args := buildComplexWhereClause(conditions)
		tx = op.apply(genericResult, tx, whereClause, args...)
	}

	genericResult.WithPassable(tx)
	return genericResult, tx.Error
}

func aggregatorFromString(aggregator string) QueryParamAggregatorEnum {
	switch aggregator {
	case QueryParamAggregatorAND.String():
		return QueryParamAggregatorAND
	case QueryParamAggregatorOR.String():
		return QueryParamAggregatorOR
	default:
		return QueryParamAggregatorNONE
	}
}

func (op *SearchInPropertiesOperator) getAllValidConditions(params QueryParams) []queryPart {
	query := parseQueryString(strings.Join(params[op.requestParamName], ","))

	validQuery := make([]queryPart, 0, len(query))
	for _, condition := range query {
		conditionSplit := strings.SplitN(condition.Query, ":", 2)
		if len(conditionSplit) == 2 && op.properties[conditionSplit[0]] != (models.SearchableProperty{}) {
			condition.Query = conditionSplit[0]
			condition.Value = conditionSplit[1]
			validQuery = append(validQuery, condition)
		}
	}

	return validQuery
}

func (op *SearchInPropertiesOperator) buildCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
	switch prop.Type {
	case "text", "varchar":
		return op.buildTextCondition(prop, searchTerm)
	case "int", "bigint":
		return op.buildIntCondition(prop, searchTerm)
	case "date":
		return op.buildDateCondition(prop, searchTerm)
	case "bool":
		return op.buildBoolCondition(prop, searchTerm)
	case "uuid":
		return op.buildUUIDCondition(prop, searchTerm)
	}
	return "", nil
}

func (op *SearchInPropertiesOperator) buildTextCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
	if prop.LikeBefore || prop.LikeAfter {
		likeTerm := searchTerm
		if prop.LikeBefore {
			likeTerm = "%" + likeTerm
		}
		if prop.LikeAfter {
			likeTerm = likeTerm + "%"
		}
		return prop.Field + " LIKE ?", likeTerm
	}
	return prop.Field + " " + prop.Comparison + " ?", searchTerm
}

func (op *SearchInPropertiesOperator) buildIntCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
	if intValue, err := strconv.Atoi(searchTerm); err == nil {
		return prop.Field + " " + prop.Comparison + " ?", intValue
	}
	return "", nil
}

func (op *SearchInPropertiesOperator) buildDateCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
	if date, err := time.Parse(prop.Format, searchTerm); err == nil {
		return prop.Field + " " + prop.Comparison + " ?", date
	}
	return "", nil
}

func (op *SearchInPropertiesOperator) buildBoolCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
	if boolValue, err := strconv.ParseBool(searchTerm); err == nil {
		return prop.Field + " = ?", boolValue
	}
	return "", nil
}

func (op *SearchInPropertiesOperator) buildUUIDCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
	if _, err := uuid.Parse(searchTerm); err == nil {
		return prop.Field + " = ?", searchTerm
	}
	return "", nil
}
