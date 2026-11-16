// Package operators is the historical home of every pipeline operator
// (SQL and Elasticsearch alike). It now acts as a re-export facade:
//
//   - The stack-agnostic primitives (Operator, Result, BaseResult, paginated
//     result types, type-ID constants, QueryParams) are defined in the
//     pipeline/operators/core subpackage and re-exported here through type
//     aliases.
//   - The concrete SQL operators were moved to pipeline/operators/sql in
//     v0.4. They are re-exported here for backwards compatibility through
//     the deprecated aliases declared in deprecated.go.
//
// New code should import pipeline/operators/sql or pipeline/operators/elastic
// directly. The aliases under operators.* will continue to work until they
// are removed in a future major version.
package operators

import "github.com/pixie-sh/database-helpers-go/pipeline/operators/core"

// Operator is re-exported from pipeline/operators/core.
type Operator = core.Operator

// QueryParams is re-exported from pipeline/operators/core.
type QueryParams = core.QueryParams

// Operator type identifiers, re-exported from pipeline/operators/core so the
// pipeline can keep matching steps on a single integer.
const (
	DatabaseOperatorType = core.DatabaseOperatorType
	ElasticOperatorType  = core.ElasticOperatorType
)
