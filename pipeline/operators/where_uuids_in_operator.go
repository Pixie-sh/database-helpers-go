package operators

import (
	"context"
	"github.com/pixie-sh/errors-go"
	pulid "github.com/pixie-sh/ulid-go"
	"strings"
)

// WhereUUIDsInOperator something amazing... or not.
type WhereUUIDsInOperator struct {
	DatabaseOperator

	property       string
	maxNumberOfIds int
}

// NewWhereUUIDsInOperator something amazing, is it?
func NewWhereUUIDsInOperator(queryParams QueryParams, property string, requestParamName string, maxNumberOfIds ...int) *WhereUUIDsInOperator {
	newOperator := &WhereUUIDsInOperator{}
	newOperator.queryParams = queryParams
	newOperator.property = property
	newOperator.requestParamName = requestParamName
	if len(maxNumberOfIds) > 0 {
		newOperator.maxNumberOfIds = maxNumberOfIds[0]
	} else {
		newOperator.maxNumberOfIds = -1
	}

	if newOperator.queryParams == nil {
		newOperator.queryParams = make(QueryParams)
	}

	// if len(newOperator.queryParams[newOperator.requestParamName]) == 0 {
	// newOperator.queryParams[newOperator.requestParamName] = []string{""}
	//}

	return newOperator
}

// Handle something amazing... who knows....
func (op *WhereUUIDsInOperator) Handle(ctx context.Context, genericResult Result) (Result, error) {
	qp, ok := op.queryParams[op.requestParamName]
	if !ok {
		return genericResult, nil
	}

	if len(qp) == 0 {
		return genericResult, nil
	}

	ids := strings.Split(qp[0], ",")
	if len(ids) == 0 || len(ids) == 1 && len(ids[0]) == 0 { //'ids=' will have a [''] which needs to be ignored
		return genericResult, nil
	}

	if op.maxNumberOfIds > 0 && len(ids) > op.maxNumberOfIds {
		return nil, errors.New("number of ids exceeds the maximum allowed")
	}

	var ulids []string
	for _, id := range ids {
		u, err :=  pulid.UnmarshalString(id)
		if err != nil {
			return nil, errors.New("invalid ulid/uuid at operator")
		}

		ulids = append(ulids, u.UUID())
	}


	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	tx = tx.Where(op.property+" IN (?)", ulids)

	genericResult.WithPassable(tx)
	return genericResult, tx.Error
}
