package operators

import "github.com/pixie-sh/database-helpers-go/pipeline/operators/core"

// Result is re-exported from pipeline/operators/core.
type Result = core.Result

// BaseResult is re-exported from pipeline/operators/core.
type BaseResult = core.BaseResult

// UntypedPaginatedResult is re-exported from pipeline/operators/core.
type UntypedPaginatedResult = core.UntypedPaginatedResult

// PaginatedResult is re-exported from pipeline/operators/core.
type PaginatedResult[T any] = core.PaginatedResult[T]

// UntypedOffsetPaginatedResult is re-exported from pipeline/operators/core.
type UntypedOffsetPaginatedResult = core.UntypedOffsetPaginatedResult

// OffsetPaginatedResult is re-exported from pipeline/operators/core.
type OffsetPaginatedResult[T any] = core.OffsetPaginatedResult[T]

// UntypedListedResult is re-exported from pipeline/operators/core.
type UntypedListedResult = core.UntypedListedResult

// ListedResult is re-exported from pipeline/operators/core.
type ListedResult[T any] = core.ListedResult[T]

// UntypedNextAfterPaginatedResult is re-exported from pipeline/operators/core.
type UntypedNextAfterPaginatedResult = core.UntypedNextAfterPaginatedResult

// NextAfterPaginatedResult is re-exported from pipeline/operators/core.
type NextAfterPaginatedResult[T any] = core.NextAfterPaginatedResult[T]

// NewResult preserves the original constructor at the historical
// pipeline/operators import path. Delegates to core.NewResult.
func NewResult(passable interface{}) *BaseResult {
	return core.NewResult(passable)
}
