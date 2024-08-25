package operators

import (
	"github.com/google/uuid"
	"github.com/pixie-sh/database-helpers-go/pipeline"
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
}

// NewGlobalSearchOperator something amazing, is it?
func NewGlobalSearchOperator(queryParams QueryParams, requestParamName string, searchableProperties ...models.SearchableProperty) *GlobalSearchOperator {
	newOperator := new(GlobalSearchOperator)
	newOperator.queryParams = queryParams
	newOperator.searchableProperties = searchableProperties
	newOperator.requestParamName = requestParamName
	return newOperator
}

func (op *GlobalSearchOperator) Handle(genericResult pipeline.Result) (pipeline.Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	searchTerm := strings.Join(op.queryParams[op.requestParamName], ",")

	query := tx

	for _, prop := range op.searchableProperties {
		condition, value := op.buildCondition(prop, searchTerm)
		if condition != "" {
			query = query.Or(condition, value)
		}
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
