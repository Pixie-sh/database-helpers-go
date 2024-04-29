package operators

import (
	"github.com/pixie-sh/database-helpers-go/pipeline"
	"github.com/pixie-sh/errors-go"
	"strconv"
)

// PaginateOperator something amazing... or not.
type PaginateOperator struct {
	DatabaseOperator

	paginationOptions []int
	dest              interface{}
}

// NewPaginateOperator something amazing, is it?
func NewPaginateOperator(queryParams QueryParams, dest interface{}, paginationOptions ...int) *PaginateOperator {
	newOperator := &PaginateOperator{}
	newOperator.dest = dest
	newOperator.paginationOptions = paginationOptions
	newOperator.queryParams = queryParams
	newOperator.predicateOverride = newOperator.predicate
	return newOperator
}

func (op *PaginateOperator) predicate() bool {
	return true
}

// GetCurrentPage something amazing... uauuuuuu
func (op *PaginateOperator) GetCurrentPage(params QueryParams) int {
	pageStrList, ok := params["page"]
	if !ok || len(pageStrList) == 0 {
		return 0
	}

	intVar, err := strconv.Atoi(pageStrList[0])
	if err != nil {
		return 0
	}

	return intVar
}

// GetCurrentLimit form query params
func (op *PaginateOperator) GetCurrentLimit(params QueryParams) int {
	pageStrList, ok := params["per_page"]
	if !ok || len(pageStrList) == 0 {
		return op.paginationOptions[0]
	}

	intVar, err := strconv.Atoi(pageStrList[0])
	if err != nil {
		return op.paginationOptions[0]
	}

	return intVar
}

// Handle something amazing... who knows....
func (op *PaginateOperator) Handle(genericResult pipeline.Result) (pipeline.Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	ctx := op.queryParams
	var paginateResult UntypedPaginatedResult

	tx.
		Offset(op.GetCurrentPage(ctx) * op.GetCurrentLimit(ctx)).
		Limit(op.GetCurrentLimit(ctx)).
		Find(op.dest).
		Offset(-1).
		Limit(-1).
		Count(&paginateResult.TotalResults)

	paginateResult.PerPage = op.GetCurrentLimit(ctx)
	paginateResult.CurrentPage = op.GetCurrentPage(ctx)
	paginateResult.AvailablePerPage = op.paginationOptions
	paginateResult.PageCount = paginateResult.TotalResults / int64(paginateResult.PerPage)

	genericResult.WithPassable(&paginateResult)
	return genericResult, tx.Error
}
