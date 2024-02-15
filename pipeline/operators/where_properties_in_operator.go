package operators

import (
	"github.com/pixie-sh/database-helpers-go/pipeline"
	"github.com/pixie-sh/errors-go"
	"strings"
)

// WherePropertiesInOperator something amazing... or not.
type WherePropertiesInOperator struct {
	DatabaseOperator

	properties []string
}

// NewWherePropertiesInOperator something amazing, is it?
func NewWherePropertiesInOperator(queryParams QueryParams, properties ...string) *WherePropertiesInOperator {
	newOperator := &WherePropertiesInOperator{}
	newOperator.properties = properties
	newOperator.queryParams = queryParams
	newOperator.predicateOverride = newOperator.predicate
	return newOperator
}

func (op *WherePropertiesInOperator) predicate() bool {
	for _, prop := range op.properties {
		if len(op.queryParams[prop]) != 0 {
			return true
		}
	}

	return false
}

// Handle something amazing... who knows....
func (op *WherePropertiesInOperator) Handle(genericResult pipeline.Result) (pipeline.Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	for _, prop := range op.properties {
		value := strings.Join(op.queryParams[prop], ",")
		if value != "" {
			tx = tx.Where(prop+" = ?", value)
		}
	}

	genericResult.WithPassable(tx)
	return genericResult, nil
}
