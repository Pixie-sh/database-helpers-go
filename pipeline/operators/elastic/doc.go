// Package elastic provides dependency-free Elasticsearch/OpenSearch search body
// builders and pipeline operators.
//
// The package intentionally builds map-based request bodies instead of binding to
// a specific Elasticsearch client. Callers can send Builder.SearchRequest().Body
// through the official Elasticsearch client, an OpenSearch client, or direct HTTP.
//
// Operators in this package use operators.ElasticOperatorType and expect the
// pipeline passable value to be an *elastic.Builder.
//
// Example:
//
//	result, err := pipeline.NewPipeline(log).
//		AddOperator(
//			elastic.NewSourceOperator("id", "title"),
//			elastic.NewFilterOperator(elastic.Term("status", "live")),
//			elastic.NewFromSizePaginationOperator(0, 30),
//		).
//		ExecWithPassable(ctx, operators.NewResult(elastic.NewBuilder("deals")))
//
// Script fragments can be composed through ScriptBuilder-backed operators and
// attached as a script_score query with NewScriptScoreOperator.
package elastic
