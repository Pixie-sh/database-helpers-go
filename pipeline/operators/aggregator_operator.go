package operators

import (
	"context"
	"github.com/pixie-sh/database-helpers-go/pipeline/operators/operator_errors"
	"github.com/pixie-sh/errors-go"
)

type AggregatorOperator struct {
	DatabaseOperator
	aggregator AggregatorOperatorEnum
}

func NewAggregatorOperator(aggregator AggregatorOperatorEnum) *AggregatorOperator {
	return &AggregatorOperator{
		DatabaseOperator: DatabaseOperator{
			predicateOverride: func() bool {
				return true
			},
		},
		aggregator: aggregator,
	}
}

func (op *AggregatorOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	typedRes, ok := genericResult.(*BaseResult)
	if !ok {
		return genericResult, errors.New("invalid result type").WithErrorCode(operator_errors.InvalidResultTypeErrorCode)
	}

	typedRes.previous = op
	return typedRes, nil
}
