package operators

import "github.com/pixie-sh/database-helpers-go/pipeline/operators/sql"

// This file re-exports every SQL operator (and related symbol) that used to
// live in pipeline/operators directly, before they were moved to
// pipeline/operators/sql in v0.4.
//
// Every alias below carries a // Deprecated: tag pointing at the new
// canonical location. Existing call sites compile unchanged; new code should
// import pipeline/operators/sql directly.

// Deprecated: use sql.DatabaseOperator instead.
type DatabaseOperator = sql.DatabaseOperator

// Deprecated: use sql.AggregatorOperator instead.
type AggregatorOperator = sql.AggregatorOperator

// Deprecated: use sql.NewAggregatorOperator instead.
var NewAggregatorOperator = sql.NewAggregatorOperator

// Deprecated: use sql.GlobalSearchOperator instead.
type GlobalSearchOperator = sql.GlobalSearchOperator

// Deprecated: use sql.NewGlobalSearchOperator instead.
var NewGlobalSearchOperator = sql.NewGlobalSearchOperator

// Deprecated: use sql.HardWhereOperator instead.
type HardWhereOperator = sql.HardWhereOperator

// Deprecated: use sql.NewHardWhereOperator instead.
var NewHardWhereOperator = sql.NewHardWhereOperator

// Deprecated: use sql.JsonSearchOperator instead.
type JsonSearchOperator = sql.JsonSearchOperator

// Deprecated: use sql.NewJsonSearchOperator instead.
var NewJsonSearchOperator = sql.NewJsonSearchOperator

// Deprecated: use sql.JsonULIDSearchOperator instead.
type JsonULIDSearchOperator = sql.JsonULIDSearchOperator

// Deprecated: use sql.NewJsonULIDSearchOperator instead.
var NewJsonULIDSearchOperator = sql.NewJsonULIDSearchOperator

// Deprecated: use sql.ListOperator instead.
type ListOperator = sql.ListOperator

// Deprecated: use sql.NewListOperator instead.
var NewListOperator = sql.NewListOperator

// Deprecated: use sql.OffsetPaginateOperator instead.
type OffsetPaginateOperator = sql.OffsetPaginateOperator

// Deprecated: use sql.NewOffsetPaginateOperator instead.
var NewOffsetPaginateOperator = sql.NewOffsetPaginateOperator

// Deprecated: use sql.OrderByOperator instead.
type OrderByOperator = sql.OrderByOperator

// Deprecated: use sql.NewOrderByOperator instead.
var NewOrderByOperator = sql.NewOrderByOperator

// Deprecated: use sql.PaginateOperator instead.
type PaginateOperator = sql.PaginateOperator

// Deprecated: use sql.NewPaginateOperator instead.
var NewPaginateOperator = sql.NewPaginateOperator

// Deprecated: use sql.SearchInPropertiesOperator instead.
type SearchInPropertiesOperator = sql.SearchInPropertiesOperator

// Deprecated: use sql.NewSearchInPropertiesOperator instead.
var NewSearchInPropertiesOperator = sql.NewSearchInPropertiesOperator

// Deprecated: use sql.WhereIdsInOperator instead.
type WhereIdsInOperator = sql.WhereIdsInOperator

// Deprecated: use sql.NewWhereIdsInOperator instead.
var NewWhereIdsInOperator = sql.NewWhereIdsInOperator

// Deprecated: use sql.WherePropertiesInOperator instead.
type WherePropertiesInOperator = sql.WherePropertiesInOperator

// Deprecated: use sql.NewWherePropertiesInOperator instead.
var NewWherePropertiesInOperator = sql.NewWherePropertiesInOperator

// Deprecated: use sql.WhereUUIDsInOperator instead.
type WhereUUIDsInOperator = sql.WhereUUIDsInOperator

// Deprecated: use sql.NewWhereUUIDsInOperator instead.
var NewWhereUUIDsInOperator = sql.NewWhereUUIDsInOperator

// Deprecated: use sql.AggregatorOperatorEnum instead.
type AggregatorOperatorEnum = sql.AggregatorOperatorEnum

// Deprecated: use sql.AggregatorConditionOR instead.
const AggregatorConditionOR = sql.AggregatorConditionOR

// Deprecated: use sql.AggregatorConditionAND instead.
const AggregatorConditionAND = sql.AggregatorConditionAND

// Deprecated: use sql.QueryParamAggregatorEnum instead.
type QueryParamAggregatorEnum = sql.QueryParamAggregatorEnum

// Deprecated: use sql.QueryParamAggregatorOR instead.
const QueryParamAggregatorOR = sql.QueryParamAggregatorOR

// Deprecated: use sql.QueryParamAggregatorAND instead.
const QueryParamAggregatorAND = sql.QueryParamAggregatorAND

// Deprecated: use sql.QueryParamAggregatorNONE instead.
const QueryParamAggregatorNONE = sql.QueryParamAggregatorNONE

// Deprecated: use sql.RemoveAccentFunction instead.
const RemoveAccentFunction = sql.RemoveAccentFunction
