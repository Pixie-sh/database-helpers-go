package sql

import (
	"context"
	"fmt"

	"github.com/pixie-sh/errors-go"
	pulid "github.com/pixie-sh/ulid-go"

	databaserrors "github.com/pixie-sh/database-helpers-go/errors"
)

// JsonULIDSearchOperator something amazing... or not.
type JsonULIDSearchOperator struct {
	DatabaseOperator

	whereClause    string
	whereFormat    string
	withUUIDString bool
}

// NewJsonULIDSearchOperator something amazing, is it? yes it is
func NewJsonULIDSearchOperator(
	queryParams QueryParams,
	requestParamName string,
	whereClause string,
	whereFormat string,
	withUUIDString bool,
) *JsonULIDSearchOperator {
	newOperator := new(JsonULIDSearchOperator)
	newOperator.queryParams = queryParams
	newOperator.whereClause = whereClause
	newOperator.whereFormat = whereFormat
	newOperator.requestParamName = requestParamName
	newOperator.withUUIDString = withUUIDString
	return newOperator
}

// Handle something amazing... who knows....
func (op *JsonULIDSearchOperator) Handle(ctx context.Context, genericResult Result) (Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(databaserrors.InvalidPassableErrorCode)
	}

	uid, err := pulid.UnmarshalString(op.queryParams[op.requestParamName][0])
	if err != nil {
		return nil, errors.NewWithError(err, "invalid uuid at operator").WithErrorCode(databaserrors.InvalidULIDErrorCode)
	}

	var uidString string
	if op.withUUIDString {
		uidString = uid.UUID()
	} else {
		uidString = uid.String()
	}

	if len(uidString) == 0 {
		return nil, errors.New("invalid uuid string at operator").WithErrorCode(databaserrors.EmptyULIDErrorCode)
	}

	genericResult.WithPassable(op.apply(genericResult, tx, op.whereClause, fmt.Sprintf(op.whereFormat, uidString)))
	return genericResult, tx.Error
}
