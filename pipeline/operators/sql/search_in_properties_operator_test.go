package sql

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pixie-sh/database-helpers-go/pipeline/operators/models"
)

const testULID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"

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
			Value:      []any{"full_time", "part_time"},
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

func TestSearchInPropertiesOperatorPreservesConditionOrder(t *testing.T) {
	operator := NewSearchInPropertiesOperator(
		QueryParams{"search_in": {"status:pending{AND}resource_id:" + testULID}},
		"search_in",
		models.SearchableProperty{Field: "status", Type: "text", Comparison: "="},
		models.SearchableProperty{Field: "resource_id", Type: "uuid", Comparison: "="},
	)

	searchTerms := operator.getAllValidConditions(operator.queryParams)
	fieldOrder, groupedTerms := groupSearchTermsByField(searchTerms)
	conditions := make([]queryCondition, 0, len(searchTerms))

	for _, fieldName := range fieldOrder {
		terms := groupedTerms[fieldName]

		condition, parsedValue := operator.buildCondition(operator.properties[fieldName], terms[0].Value)
		conditions = append(conditions, queryCondition{
			Condition:  condition,
			Value:      parsedValue,
			Aggregator: aggregatorFromString(terms[0].Aggregator),
		})
	}

	whereClause, args := buildComplexWhereClause(conditions)

	if whereClause != "(status = ? AND resource_id = ?)" {
		t.Fatalf("expected ordered where clause, got %s", whereClause)
	}
	if strings.Contains(whereClause, "?status") || strings.Contains(whereClause, "?resource_id") {
		t.Fatalf("where clause contains adjacent conditions without separator: %s", whereClause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "pending" {
		t.Fatalf("expected first arg to be status, got %v", args[0])
	}
	if !strings.HasPrefix(fmt.Sprint(args[1]), testULID) {
		t.Fatalf("expected second arg to be resource_id, got %v", args[1])
	}
}

func TestSearchInPropertiesOperatorPreservesOrderBeforeBuildingWhereClause(t *testing.T) {
	operator := NewSearchInPropertiesOperator(
		QueryParams{"search_in": {"status:pending{AND}resource_id:" + testULID}},
		"search_in",
		models.SearchableProperty{Field: "status", Type: "text", Comparison: "="},
		models.SearchableProperty{Field: "resource_id", Type: "uuid", Comparison: "="},
	)

	searchTerms := operator.getAllValidConditions(operator.queryParams)
	if len(searchTerms) != 2 {
		t.Fatalf("expected 2 parsed search terms, got %d", len(searchTerms))
	}
	if searchTerms[0].Query != "status" || aggregatorFromString(searchTerms[0].Aggregator) != QueryParamAggregatorAND {
		t.Fatalf("expected first parsed term to be status with AND aggregator, got %+v", searchTerms[0])
	}
	if searchTerms[1].Query != "resource_id" || aggregatorFromString(searchTerms[1].Aggregator) != QueryParamAggregatorNONE {
		t.Fatalf("expected second parsed term to be resource_id with no aggregator, got %+v", searchTerms[1])
	}

	whereClause, args := buildSearchInWhereClauseForTest(operator)

	if whereClause != "(status = ? AND resource_id = ?)" {
		t.Fatalf("expected source-ordered where clause, got %s", whereClause)
	}
	if strings.Contains(whereClause, "?status") || strings.Contains(whereClause, "?resource_id") {
		t.Fatalf("where clause contains adjacent conditions without separator: %s", whereClause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "pending" {
		t.Fatalf("expected first arg to be status, got %v", args[0])
	}
}

func TestBuildComplexWhereClauseEmptyConditions(t *testing.T) {
	whereClause, args := buildComplexWhereClause(nil)

	if whereClause != "" {
		t.Fatalf("expected empty where clause, got %q", whereClause)
	}
	if args != nil {
		t.Fatalf("expected nil args, got %#v", args)
	}
}

func TestSearchInPropertiesOperatorBuildWhereClauseCompatibility(t *testing.T) {
	operator := NewSearchInPropertiesOperator(
		QueryParams{},
		"search_in",
		models.SearchableProperty{Field: "status", Type: "text", Comparison: "="},
		models.SearchableProperty{Field: "resource_id", Type: "uuid", Comparison: "="},
		models.SearchableProperty{Field: "state", Type: "enum", Comparison: "="},
		models.SearchableProperty{Field: "name", Type: "text", LikeBefore: true, LikeAfter: true, Ilike: true},
		models.SearchableProperty{Field: "label", Type: "text", LikeBefore: true, LikeAfter: true, Ilike: true, Unaccent: true},
		models.SearchableProperty{Field: "score", Type: "bigint", Comparison: ">="},
		models.SearchableProperty{Field: "created_at", Type: "date", Comparison: ">=", Format: "2006-01-02"},
		models.SearchableProperty{Field: "active", Type: "bool"},
		models.SearchableProperty{Field: "tags", Type: "[]text", Comparison: "&&"},
	)

	tests := []struct {
		name       string
		searchIn   string
		wantClause string
		wantArgs   []any
	}{
		{name: "single text equality", searchIn: "status:pending", wantClause: "(status = ?)", wantArgs: []any{"pending"}},
		{name: "different fields with AND preserve input order", searchIn: "status:pending{AND}resource_id:" + testULID, wantClause: "(status = ? AND resource_id = ?)", wantArgs: []any{"pending", testULID}},
		{name: "different fields with OR preserve input order", searchIn: "status:pending{OR}score:5", wantClause: "(status = ? OR score >= ?)", wantArgs: []any{"pending", 5}},
		{name: "repeated field still builds IN condition", searchIn: "status:pending{OR}status:active", wantClause: "(status IN (?, ?))", wantArgs: []any{"pending", "active"}},
		{name: "ilike text wraps wildcards", searchIn: "name:sample", wantClause: "(name ILIKE ?)", wantArgs: []any{"%sample%"}},
		{name: "unaccent ilike wraps field and placeholder", searchIn: "label:cafe", wantClause: "(remove_accent(label) ILIKE remove_accent(?))", wantArgs: []any{"%cafe%"}},
		{name: "bigint parses integer value", searchIn: "score:10", wantClause: "(score >= ?)", wantArgs: []any{10}},
		{name: "date parses configured format", searchIn: "created_at:2026-06-09", wantClause: "(created_at >= ?)", wantArgs: []any{time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)}},
		{name: "bool parses boolean value", searchIn: "active:true", wantClause: "(active = ?)", wantArgs: []any{true}},
		{name: "text array overlap formats postgres array", searchIn: "tags:alpha, beta", wantClause: "(tags && ?)", wantArgs: []any{"{\"alpha\", \"beta\"}"}},
		{name: "unknown field is ignored", searchIn: "unknown:value{AND}status:pending", wantClause: "(status = ?)", wantArgs: []any{"pending"}},
		{name: "invalid int value is ignored", searchIn: "score:not-a-number{AND}status:pending", wantClause: "(status = ?)", wantArgs: []any{"pending"}},
		{name: "invalid uuid value is ignored", searchIn: "resource_id:not-a-uid{AND}status:pending", wantClause: "(status = ?)", wantArgs: []any{"pending"}},
		{name: "empty valid conditions returns no clause", searchIn: "unknown:value", wantClause: "", wantArgs: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operator.queryParams = QueryParams{"search_in": {tt.searchIn}}

			whereClause, args := buildSearchInWhereClauseForTest(operator)

			if whereClause != tt.wantClause {
				t.Fatalf("expected clause %q, got %q", tt.wantClause, whereClause)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("expected %d args, got %d: %#v", len(tt.wantArgs), len(args), args)
			}

			for i, wantArg := range tt.wantArgs {
				if !sameSearchInArgForTest(args[i], wantArg) {
					t.Fatalf("expected arg %d to be %#v, got %#v", i, wantArg, args[i])
				}
			}
		})
	}
}

func TestMatchingTermsForFieldPreservesRepeatedFieldOrder(t *testing.T) {
	searchTerms := []queryPart{
		{Query: "status", Value: "first", Aggregator: QueryParamAggregatorOR.String()},
		{Query: "resource_id", Value: "ignored", Aggregator: QueryParamAggregatorAND.String()},
		{Query: "status", Value: "second", Aggregator: QueryParamAggregatorNONE.String()},
	}

	fieldOrder, groupedTerms := groupSearchTermsByField(searchTerms)
	terms := groupedTerms["status"]

	if len(fieldOrder) != 2 || fieldOrder[0] != "status" || fieldOrder[1] != "resource_id" {
		t.Fatalf("expected first-seen field order to be preserved, got %+v", fieldOrder)
	}

	if len(terms) != 2 {
		t.Fatalf("expected 2 status terms, got %d", len(terms))
	}
	if terms[0].Value != "first" || terms[1].Value != "second" {
		t.Fatalf("expected status terms to preserve order, got %+v", terms)
	}
}

func TestFormatToPgArrayFormatsCommaSeparatedValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "multiple values", input: `alpha, beta`, want: `{"alpha", "beta"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatToPgArray(tt.input); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestSearchInPropertiesOperatorHelperBranches(t *testing.T) {
	operator := NewSearchInPropertiesOperator(QueryParams{}, "search_in")

	t.Run("buildCondition unknown type", func(t *testing.T) {
		condition, value := operator.buildCondition(models.SearchableProperty{Field: "unknown", Type: "unknown"}, "value")
		if condition != "" || value != nil {
			t.Fatalf("expected unknown type to return no condition, got %q %#v", condition, value)
		}
	})

	t.Run("buildTextCondition LIKE without ILIKE", func(t *testing.T) {
		condition, value := operator.buildTextCondition(
			models.SearchableProperty{Field: "name", Type: "text", LikeBefore: true, Comparison: "="},
			"sample",
		)
		if condition != "name LIKE ?" || value != "%sample" {
			t.Fatalf("unexpected LIKE condition/value: %q %#v", condition, value)
		}
	})

	t.Run("buildTextArrayCondition ANY branch", func(t *testing.T) {
		condition, value := operator.buildTextArrayCondition(
			models.SearchableProperty{Field: "tags", Type: "[]text", Comparison: "="},
			"alpha",
		)
		if condition != "? = ANY(tags)" || value != "alpha" {
			t.Fatalf("unexpected ANY condition/value: %q %#v", condition, value)
		}
	})

	t.Run("buildDateCondition invalid", func(t *testing.T) {
		condition, value := operator.buildDateCondition(
			models.SearchableProperty{Field: "created_at", Type: "date", Comparison: ">=", Format: "2006-01-02"},
			"not-a-date",
		)
		if condition != "" || value != nil {
			t.Fatalf("expected invalid date to return no condition, got %q %#v", condition, value)
		}
	})

	t.Run("buildBoolCondition invalid", func(t *testing.T) {
		condition, value := operator.buildBoolCondition(models.SearchableProperty{Field: "active", Type: "bool"}, "not-bool")
		if condition != "" || value != nil {
			t.Fatalf("expected invalid bool to return no condition, got %q %#v", condition, value)
		}
	})
}

func TestBuildInConditionBranches(t *testing.T) {
	operator := NewSearchInPropertiesOperator(QueryParams{}, "search_in")

	tests := []struct {
		name       string
		prop       models.SearchableProperty
		terms      []queryPart
		wantClause string
		wantValues []any
	}{
		{name: "varchar values", prop: models.SearchableProperty{Field: "name", Type: "varchar"}, terms: []queryPart{{Value: "alice"}, {Value: "bob"}}, wantClause: "name IN (?, ?)", wantValues: []any{"alice", "bob"}},
		{name: "enum values", prop: models.SearchableProperty{Field: "status", Type: "enum"}, terms: []queryPart{{Value: "pending"}, {Value: "approved"}}, wantClause: "status IN (?, ?)", wantValues: []any{"pending", "approved"}},
		{name: "int skips invalid values", prop: models.SearchableProperty{Field: "rating", Type: "int"}, terms: []queryPart{{Value: "bad"}, {Value: "7"}}, wantClause: "rating IN (?)", wantValues: []any{7}},
		{name: "date skips invalid values", prop: models.SearchableProperty{Field: "created_at", Type: "date", Format: "2006-01-02"}, terms: []queryPart{{Value: "bad"}, {Value: "2026-06-09"}}, wantClause: "created_at IN (?)", wantValues: []any{time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)}},
		{name: "bool skips invalid values", prop: models.SearchableProperty{Field: "active", Type: "bool"}, terms: []queryPart{{Value: "bad"}, {Value: "true"}}, wantClause: "active IN (?)", wantValues: []any{true}},
		{name: "uuid skips invalid values", prop: models.SearchableProperty{Field: "resource_id", Type: "uuid"}, terms: []queryPart{{Value: "bad"}, {Value: testULID}}, wantClause: "resource_id IN (?)", wantValues: []any{testULID}},
		{name: "unsupported type returns no condition", prop: models.SearchableProperty{Field: "unknown", Type: "unknown"}, terms: []queryPart{{Value: "a"}, {Value: "b"}}, wantClause: "", wantValues: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition, values := operator.buildInCondition(tt.prop, tt.terms)

			if condition != tt.wantClause {
				t.Fatalf("expected clause %q, got %q", tt.wantClause, condition)
			}
			if len(values) != len(tt.wantValues) {
				t.Fatalf("expected %d values, got %d: %#v", len(tt.wantValues), len(values), values)
			}

			for i, wantValue := range tt.wantValues {
				if !sameSearchInArgForTest(values[i], wantValue) {
					t.Fatalf("expected value %d to be %#v, got %#v", i, wantValue, values[i])
				}
			}
		})
	}
}

func sameSearchInArgForTest(got any, want any) bool {
	gotString := fmt.Sprint(got)
	wantString := fmt.Sprint(want)
	return gotString == wantString || strings.HasPrefix(gotString, wantString+"(")
}

func buildSearchInWhereClauseForTest(operator *SearchInPropertiesOperator) (string, []any) {
	searchTerms := operator.getAllValidConditions(operator.queryParams)
	fieldOrder, groupedTerms := groupSearchTermsByField(searchTerms)
	conditions := make([]queryCondition, 0, len(searchTerms))

	for _, fieldName := range fieldOrder {
		prop, ok := operator.properties[fieldName]
		if !ok {
			continue
		}

		terms := groupedTerms[fieldName]

		if len(terms) == 1 {
			condition, parsedValue := operator.buildCondition(prop, terms[0].Value)
			if condition != "" {
				conditions = append(conditions, queryCondition{
					Condition:  condition,
					Value:      parsedValue,
					Aggregator: aggregatorFromString(terms[0].Aggregator),
				})
			}
		} else {
			condition, parsedValues := operator.buildInCondition(prop, terms)
			if condition != "" {
				conditions = append(conditions, queryCondition{
					Condition:  condition,
					Value:      parsedValues,
					Aggregator: aggregatorFromString(terms[0].Aggregator),
				})
			}
		}
	}

	return buildComplexWhereClause(conditions)
}
