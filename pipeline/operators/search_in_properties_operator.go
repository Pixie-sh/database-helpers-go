package operators

import (
	"context"
	"fmt"

	"github.com/pixie-sh/database-helpers-go/pipeline/operators/models"
	"github.com/pixie-sh/errors-go"
	pulid "github.com/pixie-sh/ulid-go"
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

	// Group search terms by field
	groupedTerms := make(map[string][]queryPart)
	for _, term := range searchTerms {
		groupedTerms[term.Query] = append(groupedTerms[term.Query], term)
	}

	var conditions []queryCondition

	for fieldName, terms := range groupedTerms {
		if prop, ok := op.properties[fieldName]; ok {
			if len(terms) == 1 {
				// Single value: build normal condition
				condition, parsedValue := op.buildCondition(prop, terms[0].Value)
				if condition != "" {
					conditions = append(conditions, queryCondition{
						Condition:  condition,
						Value:      parsedValue,
						Aggregator: aggregatorFromString(terms[0].Aggregator),
					})
				}
			} else {
				// Multiple values: build IN condition
				condition, parsedValues := op.buildInCondition(prop, terms)
				if condition != "" {
					conditions = append(conditions, queryCondition{
						Condition:  condition,
						Value:      parsedValues,
						Aggregator: aggregatorFromString(terms[0].Aggregator),
					})
				}
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

func (op *SearchInPropertiesOperator) buildInCondition(prop models.SearchableProperty, terms []queryPart) (string, []interface{}) {
	var values []interface{}
	
	for _, term := range terms {
		var parsedValue interface{}
		
		switch prop.Type {
		case "text", "varchar", "enum":
			parsedValue = term.Value
		case "int", "bigint":
			if intValue, err := strconv.Atoi(term.Value); err == nil {
				parsedValue = intValue
			}
		case "date":
			if date, err := time.Parse(prop.Format, term.Value); err == nil {
				parsedValue = date
			}
		case "bool":
			if boolValue, err := strconv.ParseBool(term.Value); err == nil {
				parsedValue = boolValue
			}
		case "uuid":
			if ulid, err := pulid.UnmarshalString(term.Value); err == nil {
				parsedValue = ulid
			}
		}
		
		if parsedValue != nil {
			values = append(values, parsedValue)
		}
	}
	
	if len(values) == 0 {
		return "", nil
	}
	
	// Build placeholders for IN clause
	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	
	return fmt.Sprintf("%s IN (%s)", prop.Field, strings.Join(placeholders, ", ")), values
}

func (op *SearchInPropertiesOperator) buildTextCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
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
	if ulid, err := pulid.UnmarshalString(searchTerm); err == nil {
		return prop.Field + " = ?", ulid
	}
	return "", nil
}
