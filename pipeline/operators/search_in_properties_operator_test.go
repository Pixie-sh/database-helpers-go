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

func TestBuildGroupedConditionsCanonical(t *testing.T) {
	// Integration test: canonical query from spec
	// {P}functional_area_01:Logística{|}functional_area_01:Putas{/P}{&}district:Beja
	// Expected SQL: (functional_area_01 = ? OR functional_area_01 = ?) AND district = ?

	operator := NewSearchInPropertiesOperator(
		QueryParams{},
		"search",
		models.SearchableProperty{
			Field:      "functional_area_01",
			Type:       "text",
			Comparison: "=",
		},
		models.SearchableProperty{
			Field:      "district",
			Type:       "text",
			Comparison: "=",
		},
	)

	input := "{P}functional_area_01:Logística{|}functional_area_01:Putas{/P}{&}district:Beja"
	nodes := parseQueryStringGrouped(input)

	if len(nodes) != 2 {
		t.Fatalf("expected 2 top-level nodes, got %d", len(nodes))
	}

	groupedConds := operator.buildGroupedConditions(nodes)

	if len(groupedConds) != 2 {
		t.Fatalf("expected 2 grouped conditions, got %d", len(groupedConds))
	}

	whereClause, args := buildGroupedWhereClause(groupedConds)

	expectedSQL := "(functional_area_01 = ? OR functional_area_01 = ?) AND district = ?"
	if whereClause != expectedSQL {
		t.Errorf("SQL mismatch:\n  want: %q\n  got:  %q", expectedSQL, whereClause)
	}

	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(args))
	}

	if args[0] != "Logística" {
		t.Errorf("arg[0]: want 'Logística', got '%v'", args[0])
	}
	if args[1] != "Putas" {
		t.Errorf("arg[1]: want 'Putas', got '%v'", args[1])
	}
	if args[2] != "Beja" {
		t.Errorf("arg[2]: want 'Beja', got '%v'", args[2])
	}
}

func TestBuildGroupedConditionsInvalidFieldStripped(t *testing.T) {
	// Test that invalid fields are silently stripped in grouped mode
	operator := NewSearchInPropertiesOperator(
		QueryParams{},
		"search",
		models.SearchableProperty{
			Field:      "district",
			Type:       "text",
			Comparison: "=",
		},
	)

	// functional_area_01 is NOT in the property whitelist
	input := "{P}functional_area_01:Logística{|}functional_area_01:Putas{/P}{&}district:Beja"
	nodes := parseQueryStringGrouped(input)
	groupedConds := operator.buildGroupedConditions(nodes)

	// The group should be empty (invalid fields stripped), only leaf remains
	if len(groupedConds) != 1 {
		t.Fatalf("expected 1 grouped condition (group stripped), got %d", len(groupedConds))
	}

	whereClause, args := buildGroupedWhereClause(groupedConds)
	if whereClause != "district = ?" {
		t.Errorf("expected 'district = ?', got '%s'", whereClause)
	}
	if len(args) != 1 || args[0] != "Beja" {
		t.Errorf("expected args ['Beja'], got %v", args)
	}
}

func TestBuildGroupedConditionsMultipleGroupsWithAND(t *testing.T) {
	operator := NewSearchInPropertiesOperator(
		QueryParams{},
		"search",
		models.SearchableProperty{
			Field:      "f1",
			Type:       "text",
			Comparison: "=",
		},
		models.SearchableProperty{
			Field:      "f2",
			Type:       "text",
			Comparison: "=",
		},
	)

	input := "{P}f1:v1{|}f1:v2{/P}{&}{P}f2:v3{|}f2:v4{/P}"
	nodes := parseQueryStringGrouped(input)
	groupedConds := operator.buildGroupedConditions(nodes)

	whereClause, args := buildGroupedWhereClause(groupedConds)

	expectedSQL := "(f1 = ? OR f1 = ?) AND (f2 = ? OR f2 = ?)"
	if whereClause != expectedSQL {
		t.Errorf("SQL mismatch:\n  want: %q\n  got:  %q", expectedSQL, whereClause)
	}
	if len(args) != 4 {
		t.Errorf("expected 4 args, got %d", len(args))
	}
}

func TestFlatPathBackwardCompat(t *testing.T) {
	// Verify that flat queries (no group tokens) still work identically
	// by testing the operator's getAllValidConditions path

	operator := NewSearchInPropertiesOperator(
		QueryParams{
			"search": []string{"field1:val1{AND}field2:val2"},
		},
		"search",
		models.SearchableProperty{
			Field:      "field1",
			Type:       "text",
			Comparison: "=",
		},
		models.SearchableProperty{
			Field:      "field2",
			Type:       "text",
			Comparison: "=",
		},
	)

	terms := operator.getAllValidConditions(operator.queryParams)
	if len(terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(terms))
	}
	if terms[0].Query != "field1" || terms[0].Value != "val1" {
		t.Errorf("term[0]: want field1:val1, got %s:%s", terms[0].Query, terms[0].Value)
	}
	if terms[1].Query != "field2" || terms[1].Value != "val2" {
		t.Errorf("term[1]: want field2:val2, got %s:%s", terms[1].Query, terms[1].Value)
	}
}

func TestGroupedDeterminismStress(t *testing.T) {
	// 100 iterations of the same grouped query through the full condition building pipeline
	operator := NewSearchInPropertiesOperator(
		QueryParams{},
		"search",
		models.SearchableProperty{
			Field:      "area",
			Type:       "text",
			Comparison: "=",
		},
		models.SearchableProperty{
			Field:      "district",
			Type:       "text",
			Comparison: "=",
		},
		models.SearchableProperty{
			Field:      "status",
			Type:       "text",
			Comparison: "=",
		},
	)

	input := "{P}area:A{|}area:B{/P}{&}district:C{&}status:D"
	var referenceSQL string
	var referenceArgCount int

	for i := 0; i < 100; i++ {
		nodes := parseQueryStringGrouped(input)
		groupedConds := operator.buildGroupedConditions(nodes)
		sql, args := buildGroupedWhereClause(groupedConds)

		if i == 0 {
			referenceSQL = sql
			referenceArgCount = len(args)
			continue
		}

		if sql != referenceSQL {
			t.Fatalf("iteration %d: SQL changed: %q vs %q", i, sql, referenceSQL)
		}
		if len(args) != referenceArgCount {
			t.Fatalf("iteration %d: arg count changed: %d vs %d", i, len(args), referenceArgCount)
		}
	}
}
