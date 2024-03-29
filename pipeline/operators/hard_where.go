package operators

import (
	"github.com/pixie-sh/database-helpers-go/pipeline"
	"github.com/pixie-sh/errors-go"
)

// HardWhereOperator this where will always be present no matter the request
type HardWhereOperator struct {
	DatabaseOperator

	prop  string
	value any
}

// NewHardWhereOperator something amazing, is it? idk, but its the same as the above
func NewHardWhereOperator(prop string, value any) *HardWhereOperator {
	newOperator := &HardWhereOperator{}
	newOperator.prop = prop
	newOperator.value = value
	newOperator.predicateOverride = newOperator.predicate
	return newOperator
}

// Predicate override so it's always called and handled
func (op *HardWhereOperator) predicate() bool {
	return true
}

// Handle make the magic happen
func (op *HardWhereOperator) Handle(genericResult pipeline.Result) (pipeline.Result, error) {
	tx, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable")
	}

	tx = tx.Where(op.prop, op.value)

	genericResult.WithPassable(tx)
	return genericResult, tx.Error
}
