package operators

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pixie-sh/database-helpers-go/pipeline/operators/models"
	"github.com/pixie-sh/errors-go"
	"strconv"
	"strings"
	"time"
)

// GlobalSearchOperator something amazing... or not.
type GlobalSearchOperator struct {
	DatabaseOperator

	searchableProperties []models.SearchableProperty
	aggregatorCondition  AggregatorOperatorEnum
}

// NewGlobalSearchOperator something amazing, is it?
func NewGlobalSearchOperator(queryParams QueryParams, requestParamName string, searchableProperties ...models.SearchableProperty) *GlobalSearchOperator {
	newOperator := &GlobalSearchOperator{}
	newOperator.queryParams = queryParams
	newOperator.searchableProperties = searchableProperties
	newOperator.requestParamName = requestParamName
	newOperator.aggregatorCondition = AggregatorConditionOR

	return newOperator
}

func (op *GlobalSearchOperator) WithAggregatorCondition(condition AggregatorOperatorEnum) *GlobalSearchOperator {
	op.aggregatorCondition = condition
	return op
}

func (op *GlobalSearchOperator) Handle(ctx context.Context, genericResult Result) (Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	searchTerm := strings.Join(op.queryParams[op.requestParamName], ",")

	query := tx

	var conditions []string
	var orValues []interface{}

	for _, prop := range op.searchableProperties {
		condition, value := op.buildCondition(prop, searchTerm)
		if condition != "" {
			conditions = append(conditions, condition)
			orValues = append(orValues, value)
		}
	}

	if len(conditions) > 0 {
		whereClause := "(" + strings.Join(conditions, op.aggregatorCondition.String()) + ")"
		query = op.apply(genericResult, query, whereClause, orValues...)
	}

	genericResult.WithPassable(query)
	return genericResult, query.Error
}

func (op *GlobalSearchOperator) buildCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
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

func (op *GlobalSearchOperator) buildTextCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
	if prop.LikeBefore || prop.LikeAfter || prop.Ilike || prop.Unaccent {
		var likeTerm string
		var fieldTerm string
		var likeOperator string

		if prop.Ilike {
			likeOperator = "ILIKE"
		} else {
			likeOperator = "LIKE"
		}

		likeTerm = searchTerm
		if prop.LikeBefore {
			likeTerm = "%" + likeTerm
		}
		if prop.LikeAfter {
			likeTerm = likeTerm + "%"
		}

		if prop.Unaccent {
			likeOperator = fmt.Sprintf(" %s remove_accent(?)", likeOperator)
			fieldTerm = fmt.Sprintf("remove_accent(%s)", prop.Field)
		} else {
			likeOperator = fmt.Sprintf(" %s ?", likeOperator)
			fieldTerm = prop.Field
		}
		return fieldTerm + likeOperator, likeTerm
	}

	return prop.Field + " " + prop.Comparison + " ?", searchTerm
}

func (op *GlobalSearchOperator) buildIntCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
	if intValue, err := strconv.Atoi(searchTerm); err == nil {
		return prop.Field + " " + prop.Comparison + " ?", intValue
	}
	return "", nil
}

func (op *GlobalSearchOperator) buildDateCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
	if date, err := time.Parse(prop.Format, searchTerm); err == nil {
		return prop.Field + " " + prop.Comparison + " ?", date
	}
	return "", nil
}

func (op *GlobalSearchOperator) buildBoolCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
	if boolValue, err := strconv.ParseBool(searchTerm); err == nil {
		return prop.Field + " = ?", boolValue
	}
	return "", nil
}

func (op *GlobalSearchOperator) buildUUIDCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
	if _, err := uuid.Parse(searchTerm); err == nil {
		return prop.Field + " = ?", searchTerm
	}
	return "", nil
}
