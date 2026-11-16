package sql

import (
	"context"
	"fmt"
	"strings"

	databaserrors "github.com/pixie-sh/database-helpers-go/errors"
	"github.com/pixie-sh/errors-go"
)

// JsonSearchOperator something amazing... or not.
type JsonSearchOperator struct {
	DatabaseOperator

	whereClause string
	whereFormat string
}

// NewJsonSearchOperator something amazing, is it?
func NewJsonSearchOperator(
	queryParams QueryParams,
	requestParamName string,
	whereClause string,
	whereFormat string,
) *JsonSearchOperator {
	newOperator := new(JsonSearchOperator)
	newOperator.queryParams = queryParams
	newOperator.whereClause = whereClause
	newOperator.whereFormat = whereFormat
	newOperator.requestParamName = requestParamName
	return newOperator
}

// Handle something amazing... who knows....
func (op *JsonSearchOperator) Handle(ctx context.Context, genericResult Result) (Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(databaserrors.InvalidPassableErrorCode)
	}

	genericResult.WithPassable(op.apply(genericResult, tx, op.whereClause, fmt.Sprintf(op.whereFormat, strings.Join(op.queryParams[op.requestParamName], ","))))
	return genericResult, tx.Error
}
