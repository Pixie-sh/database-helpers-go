package sql

import "strings"

// RemoveAccentFunction is the Postgres SQL definition of a helper function
// used by GlobalSearchOperator / SearchInPropertiesOperator when a
// SearchableProperty requests accent-insensitive matching.
const RemoveAccentFunction = `
CREATE OR REPLACE FUNCTION remove_accent(text) RETURNS text AS $$
SELECT translate($1,
    'áàâãäåāăąÁÀÂÃÄÅĀĂĄéèêëēĕėęěÉÈÊËĒĔĖĘĚíìîïīĭįİÍÌÎÏĪĬĮİóòôõöōŏőÓÒÔÕÖŌŎŐúùûüūŭůűųÚÙÛÜŪŬŮŰŲ',
    'aaaaaaaaaAAAAAAAAeeeeeeeeeeEEEEEEEEEiiiiiiiIIIIIIIIoooooooooOOOOOOOOOuuuuuuuuuUUUUUUUUU'
);
$$ LANGUAGE SQL IMMUTABLE STRICT;`

// QueryParamAggregatorEnum is the {AND}/{OR}/{-} marker used in the
// SearchInPropertiesOperator query-param syntax.
type QueryParamAggregatorEnum string

func (e QueryParamAggregatorEnum) String() string {
	return string(e)
}

const QueryParamAggregatorOR QueryParamAggregatorEnum = "{OR}"
const QueryParamAggregatorAND QueryParamAggregatorEnum = "{AND}"
const QueryParamAggregatorNONE QueryParamAggregatorEnum = "{-}"

// AggregatorOperatorEnum is the literal SQL boolean combinator (note the
// surrounding spaces) used by AggregatorOperator / DatabaseOperator.apply.
type AggregatorOperatorEnum string

func (enum AggregatorOperatorEnum) String() string {
	return string(enum)
}

const AggregatorConditionOR AggregatorOperatorEnum = " OR "
const AggregatorConditionAND AggregatorOperatorEnum = " AND "

type queryPart struct {
	Query      string
	Value      string
	Aggregator string
}

type queryCondition struct {
	Condition  string
	Value      interface{}
	Aggregator QueryParamAggregatorEnum
}

func parseQueryString(input string) []queryPart {
	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == '{' || r == '}'
	})

	var result []queryPart

	for i, part := range parts {
		if strings.Contains(part, ":") {
			query := strings.TrimSpace(part)
			var aggregator string
			if i+1 < len(parts) {
				aggregator = "{" + parts[i+1] + "}"
			}

			result = append(result, queryPart{Query: query, Aggregator: aggregator})
		}
	}

	return result
}

func buildComplexWhereClause(conditions []queryCondition) (string, []interface{}) {
	if len(conditions) == 0 {
		return "", nil
	}

	var fullCondition strings.Builder
	var args []interface{}

	fullCondition.WriteString("(")

	nextAggregator := QueryParamAggregatorAND
	for i, cond := range conditions {
		if i == 0 {
			nextAggregator = cond.Aggregator
		} else {
			switch nextAggregator {
			case QueryParamAggregatorOR:
				fullCondition.WriteString(AggregatorConditionOR.String())
				nextAggregator = cond.Aggregator
			case QueryParamAggregatorAND:
				fullCondition.WriteString(AggregatorConditionAND.String())
				nextAggregator = cond.Aggregator
			case QueryParamAggregatorNONE:
			default:
			}
		}

		fullCondition.WriteString(cond.Condition)

		// Handle both single values and slices (for IN clauses)
		if valueSlice, ok := cond.Value.([]interface{}); ok {
			args = append(args, valueSlice...)
		} else {
			args = append(args, cond.Value)
		}
	}

	fullCondition.WriteString(")")

	return fullCondition.String(), args
}
