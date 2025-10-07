package operators

import (
	"context"
	"github.com/pixie-sh/errors-go"
	"strconv"
)

// PaginateOperator something amazing... or not.
type PaginateOperator struct {
	DatabaseOperator

	paginationOptions []int
	dest              interface{}
	usePluck          bool
	pluckColumn       string
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

func (op *PaginateOperator) UsePluck(pluckColumn string) *PaginateOperator {
	op.usePluck = true
	op.pluckColumn = pluckColumn
	return op
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
func (op *PaginateOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	ctx := op.queryParams
	var paginateResult UntypedPaginatedResult

	tx.
		Offset(op.GetCurrentPage(ctx) * op.GetCurrentLimit(ctx)).
		Limit(op.GetCurrentLimit(ctx))

	if op.usePluck {
		tx.Pluck(op.pluckColumn, op.dest)
	} else {
		tx.Find(op.dest)
	}

	tx.
		Offset(-1).
		Limit(-1).
		Count(&paginateResult.TotalResults)

	paginateResult.PerPage = op.GetCurrentLimit(ctx)
	paginateResult.CurrentPage = op.GetCurrentPage(ctx)
	paginateResult.AvailablePerPage = op.paginationOptions
	paginateResult.QueryParams = op.queryParams

	if paginateResult.TotalResults != 0 {
		paginateResult.PageCount = (paginateResult.TotalResults + int64(paginateResult.PerPage) - 1) / int64(paginateResult.PerPage)
	} else {
		paginateResult.PageCount = 0
	}

	genericResult.WithPassable(&paginateResult)
	return genericResult, tx.Error
}
