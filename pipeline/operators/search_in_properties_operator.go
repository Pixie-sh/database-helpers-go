package operators

import (
	"context"
	"fmt"

	"strconv"
	"strings"
	"time"

	"github.com/pixie-sh/database-helpers-go/pipeline/operators/models"
	"github.com/pixie-sh/errors-go"
	pulid "github.com/pixie-sh/ulid-go"
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

	rawInput := strings.Join(op.queryParams[op.requestParamName], ",")

	if hasGroupTokens(rawInput) {
		// Grouped path: parse AST, validate, build grouped conditions, generate SQL
		nodes := parseQueryStringGrouped(rawInput)
		groupedConds := op.buildGroupedConditions(nodes)
		if len(groupedConds) > 0 {
			whereClause, args := buildGroupedWhereClause(groupedConds)
			if whereClause != "" {
				tx = op.apply(genericResult, tx, whereClause, args...)
			}
		}
	} else {
		// Flat path: original behavior (backward compatible)
		searchTerms := op.getAllValidConditions(op.queryParams)

		// Group search terms by field using ordered approach
		type fieldGroup struct {
			fieldName string
			terms     []queryPart
		}
		var orderedGroups []fieldGroup
		fieldIndex := make(map[string]int)

		for _, term := range searchTerms {
			if idx, ok := fieldIndex[term.Query]; ok {
				orderedGroups[idx].terms = append(orderedGroups[idx].terms, term)
			} else {
				fieldIndex[term.Query] = len(orderedGroups)
				orderedGroups = append(orderedGroups, fieldGroup{
					fieldName: term.Query,
					terms:     []queryPart{term},
				})
			}
		}

		var conditions []queryCondition

		for _, fg := range orderedGroups {
			if prop, ok := op.properties[fg.fieldName]; ok {
				if len(fg.terms) == 1 {
					condition, parsedValue := op.buildCondition(prop, fg.terms[0].Value)
					if condition != "" {
						conditions = append(conditions, queryCondition{
							Condition:  condition,
							Value:      parsedValue,
							Aggregator: aggregatorFromString(fg.terms[0].Aggregator),
						})
					}
				} else {
					condition, parsedValues := op.buildInCondition(prop, fg.terms)
					if condition != "" {
						conditions = append(conditions, queryCondition{
							Condition:  condition,
							Value:      parsedValues,
							Aggregator: aggregatorFromString(fg.terms[0].Aggregator),
						})
					}
				}
			}
		}

		if len(conditions) > 0 {
			whereClause, args := buildComplexWhereClause(conditions)
			tx = op.apply(genericResult, tx, whereClause, args...)
		}
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

// buildGroupedConditions recursively validates and converts AST queryNodes into groupedConditions.
// Each leaf is validated against the property whitelist and has its condition built.
// Groups are recursively processed. Invalid fields are silently stripped.
func (op *SearchInPropertiesOperator) buildGroupedConditions(nodes []queryNode) []groupedCondition {
	var result []groupedCondition

	for _, node := range nodes {
		switch node.Type {
		case queryNodeGroup:
			children := op.buildGroupedConditions(node.Children)
			if len(children) > 0 {
				result = append(result, groupedCondition{
					Type:       queryNodeGroup,
					Children:   children,
					Aggregator: aggregatorFromString(node.Aggregator),
				})
			}
		case queryNodeLeaf:
			conditionSplit := strings.SplitN(node.Part.Query, ":", 2)
			if len(conditionSplit) != 2 {
				continue
			}
			fieldName := conditionSplit[0]
			value := conditionSplit[1]

			prop, ok := op.properties[fieldName]
			if !ok || prop == (models.SearchableProperty{}) {
				continue
			}

			condition, parsedValue := op.buildCondition(prop, value)
			if condition == "" {
				continue
			}

			result = append(result, groupedCondition{
				Type: queryNodeLeaf,
				Condition: queryCondition{
					Condition:  condition,
					Value:      parsedValue,
					Aggregator: aggregatorFromString(node.Aggregator),
				},
				Aggregator: aggregatorFromString(node.Aggregator),
			})
		}
	}

	return result
}

func (op *SearchInPropertiesOperator) buildCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {
	switch prop.Type {
	case "text", "varchar":
		return op.buildTextCondition(prop, searchTerm)
	case "[]text":
		return op.buildTextArrayCondition(prop, searchTerm)
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

func (op *SearchInPropertiesOperator) buildTextArrayCondition(prop models.SearchableProperty, searchTerm string) (string, interface{}) {

	// Contains, string has to change
	if prop.Comparison == "@>" || prop.Comparison == "&&" {
		containsTerm := formatToPgArray(searchTerm)

		return prop.Field + " " + prop.Comparison + " ?", containsTerm
	}

	fieldTerm := fmt.Sprintf("ANY(%s)", prop.Field)
	return "? " + prop.Comparison + " " + fieldTerm, searchTerm
}

func formatToPgArray(input string) string {
	// 1. Split the string by commas
	elements := strings.Split(input, ",")

	// 2. Trim whitespace and wrap each element in double quotes
	var quoted []string
	for _, el := range elements {
		trimmed := strings.TrimSpace(el)
		quoted = append(quoted, fmt.Sprintf("\"%s\"", trimmed))
	}

	// 3. Join with commas and wrap in curly braces
	return fmt.Sprintf("{%s}", strings.Join(quoted, ", "))
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
