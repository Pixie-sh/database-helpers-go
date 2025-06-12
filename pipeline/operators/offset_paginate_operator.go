package operators

import (
	"context"
	"github.com/pixie-sh/errors-go"
	"reflect"
	"strconv"
)

// OffsetPaginateOperator something amazing... or not.
type OffsetPaginateOperator struct {
	DatabaseOperator

	paginationOptions []int
	dest              interface{}
	usePluck          bool
	pluckColumn       string
}

// NewOffsetPaginateOperator something amazing, is it?
func NewOffsetPaginateOperator(queryParams QueryParams, dest interface{}, paginationOptions ...int) *OffsetPaginateOperator {
	newOperator := &OffsetPaginateOperator{}
	newOperator.dest = dest
	newOperator.paginationOptions = paginationOptions
	newOperator.queryParams = queryParams
	newOperator.predicateOverride = newOperator.predicate
	return newOperator
}

func (op *OffsetPaginateOperator) UsePluck(pluckColumn string) *OffsetPaginateOperator {
	op.usePluck = true
	op.pluckColumn = pluckColumn
	return op
}

func (op *OffsetPaginateOperator) predicate() bool {
	return true
}

// GetCurrentPage something amazing... uauuuuuu
func (op *OffsetPaginateOperator) GetCurrentPage(params QueryParams) int {
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
func (op *OffsetPaginateOperator) GetCurrentLimit(params QueryParams) int {
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
func (op *OffsetPaginateOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	ctx := op.queryParams
	var paginateResult UntypedOffsetPaginatedResult

	tx.
		Offset(op.GetCurrentPage(ctx) * op.GetCurrentLimit(ctx)).
		Limit(op.GetCurrentLimit(ctx) + 1)

	if op.usePluck {
		tx.Pluck(op.pluckColumn, op.dest)
	} else {
		tx.Find(op.dest)
	}

	paginateResult.PerPage = op.GetCurrentLimit(ctx)
	paginateResult.CurrentPage = op.GetCurrentPage(ctx)
	paginateResult.AvailablePerPage = op.paginationOptions
	paginateResult.QueryParams = op.queryParams

	if op.dest != nil {
		destValue := reflect.ValueOf(op.dest)

		if destValue.Kind() == reflect.Ptr {
			destValue = destValue.Elem()
		}

		if destValue.Kind() == reflect.Slice {
			sliceLen := destValue.Len()
			currentLimit := op.GetCurrentLimit(ctx)

			if sliceLen > currentLimit {
				paginateResult.HasMore = true
				op.dest = destValue.Slice(0, sliceLen-1).Interface()
			}
		}
	}

	genericResult.WithPassable(&paginateResult)
	return genericResult, tx.Error
}
