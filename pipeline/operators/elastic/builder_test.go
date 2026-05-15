package elastic

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/pixie-sh/database-helpers-go/pipeline"
	base "github.com/pixie-sh/database-helpers-go/pipeline/operators"
	"github.com/pixie-sh/logger-go/logger"
)

type testLogger struct{}

func (l testLogger) Clone() logger.Interface                    { return l }
func (l testLogger) WithCtx(_ context.Context) logger.Interface { return l }
func (l testLogger) With(_ string, _ any) logger.Interface      { return l }
func (l testLogger) Log(_ string, _ ...any)                     {}
func (l testLogger) Debug(_ string, _ ...any)                   {}
func (l testLogger) Warn(_ string, _ ...any)                    {}
func (l testLogger) Error(_ string, _ ...any)                   {}

func TestBuildsElasticRequest(t *testing.T) {
	res, err := pipeline.NewPipeline(testLogger{}).
		AddOperator(
			NewSourceOperator("id", "title"),
			NewFromSizePaginationOperator(0, 30),
			NewFilterOperator(Term("status", "live")),
			NewFilterOperator(BoolShould(1,
				Term("deal_type", "online"),
				BoolFilter(
					Term("deal_type", "physical"),
					GeoDistance("locations.coords", "250km", 52.3676, 4.9041),
				),
			)),
			NewMustNotOperator(IDs("already_applied_deal_id_1", "hidden_deal_id_2")),
			NewShouldOperator(MultiMatch("laundry cleaning", "title^3", "description")),
			NewMinimumShouldMatchOperator(1),
		).
		ExecWithPassable(context.Background(), base.NewResult(NewBuilder("deals")))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	builder := res.GetPassable().(*Builder)
	body := builder.Body()
	if body["size"] != 30 {
		t.Fatalf("expected size 30, got %v", body["size"])
	}

	query := body["query"].(Query)
	boolQuery := query["bool"].(Query)
	if boolQuery["minimum_should_match"] != 1 {
		t.Fatalf("expected minimum_should_match 1, got %v", boolQuery["minimum_should_match"])
	}

	filters := boolQuery["filter"].([]interface{})
	if len(filters) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(filters))
	}

	mustNot := boolQuery["must_not"].([]interface{})
	if len(mustNot) != 1 {
		t.Fatalf("expected 1 must_not query, got %d", len(mustNot))
	}
}

func TestBuildsScriptScore(t *testing.T) {
	builder := NewBuilder("deals")
	builder.AddFilter(Term("status", "live"))

	script := builder.ScriptBuilder()
	script.Add(ScriptSectionHelpers, "double getVal(String field) { return 0.0; }")
	script.Add(ScriptSectionScoring, "score += params.w_freshness;")
	if err := script.AddParams(map[string]interface{}{"w_freshness": 25.0}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := builder.ApplyScriptScore(nil, script); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	encoded, err := json.Marshal(builder.Body())
	if err != nil {
		t.Fatalf("unexpected marshal error: %s", err)
	}
	body := string(encoded)
	if !strings.Contains(body, "script_score") {
		t.Fatalf("expected script_score in body: %s", body)
	}
	if !strings.Contains(body, "w_freshness") {
		t.Fatalf("expected script params in body: %s", body)
	}
}

func TestDuplicateScriptParamsReturnError(t *testing.T) {
	script := NewScriptBuilder()
	if err := script.AddParams(map[string]interface{}{"weight": 1}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := script.AddParams(map[string]interface{}{"weight": 2}); err == nil {
		t.Fatal("expected duplicate param error")
	}
}

func TestSearchAfterPagination(t *testing.T) {
	builder := NewBuilder("deals")
	op := NewSearchAfterPaginationOperator(
		31,
		[]Query{SortField("_score", "desc"), SortField("id.keyword", "asc")},
		12.4,
		"deal_123",
	)

	res := base.NewResult(builder)
	if _, err := op.Handle(context.Background(), res); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if builder.Body()["size"] != 31 {
		t.Fatalf("expected size 31, got %v", builder.Body()["size"])
	}
	if len(builder.Body()["search_after"].([]interface{})) != 2 {
		t.Fatalf("expected search_after values")
	}
}

func TestDebugStringIncludesRequestLine(t *testing.T) {
	builder := NewBuilder("deals")
	builder.SetSize(30)

	debug, err := builder.DebugString(true)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !strings.HasPrefix(debug, "POST /deals/_search") {
		t.Fatalf("expected request line, got %s", debug)
	}
}
