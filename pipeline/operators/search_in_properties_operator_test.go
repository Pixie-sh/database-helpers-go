package operators

import (
	"testing"

	"github.com/pixie-sh/database-helpers-go/pipeline/operators/models"
)

func TestBuildInCondition(t *testing.T) {
	operator := NewSearchInPropertiesOperator(
		QueryParams{},
		"search",
		models.SearchableProperty{
			Field: "contract_details",
			Type:  "text",
		},
	)

	terms := []queryPart{
		{Query: "contract_details", Value: "full_time", Aggregator: "{AND}"},
		{Query: "contract_details", Value: "part_time", Aggregator: ""},
	}

	condition, values := operator.buildInCondition(
		models.SearchableProperty{Field: "contract_details", Type: "text"},
		terms,
	)

	expectedCondition := "contract_details IN (?, ?)"
	if condition != expectedCondition {
		t.Errorf("Expected condition '%s', got '%s'", expectedCondition, condition)
	}

	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}

	if values[0] != "full_time" {
		t.Errorf("Expected first value 'full_time', got '%v'", values[0])
	}

	if values[1] != "part_time" {
		t.Errorf("Expected second value 'part_time', got '%v'", values[1])
	}
}

func TestBuildComplexWhereClauseWithInCondition(t *testing.T) {
	conditions := []queryCondition{
		{
			Condition:  "contract_details IN (?, ?)",
			Value:      []interface{}{"full_time", "part_time"},
			Aggregator: QueryParamAggregatorAND,
		},
	}

	whereClause, args := buildComplexWhereClause(conditions)

	expectedClause := "(contract_details IN (?, ?))"
	if whereClause != expectedClause {
		t.Errorf("Expected clause '%s', got '%s'", expectedClause, whereClause)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	if args[0] != "full_time" {
		t.Errorf("Expected first arg 'full_time', got '%v'", args[0])
	}

	if args[1] != "part_time" {
		t.Errorf("Expected second arg 'part_time', got '%v'", args[1])
	}
}
