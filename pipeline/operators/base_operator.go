package operators

import (
	"github.com/pixie-sh/database-helpers-go/database"
	"github.com/pixie-sh/database-helpers-go/pipeline"
	"github.com/pixie-sh/errors-go"
	"reflect"
)

// define the operators types
// this is used to keep the operators compatible for the whole pipeline
const (
	// DatabaseOperatorType expect that the result is a *database.DB
	// any other format shall be implemented in other operator type.
	DatabaseOperatorType uint = 0
)

// QueryParams to be used with all DatabaseOperator
type QueryParams = map[string][]string

// DatabaseOperator simple struct to hold common DatabaseOperator functions
type DatabaseOperator struct {
	predicateOverride func() bool
	queryParams       QueryParams
	requestParamName  string
}

// GetType return type DatabaseOperatorType
func (b *DatabaseOperator) GetType() uint {
	return DatabaseOperatorType
}

// Predicate basic predicate
func (b *DatabaseOperator) Predicate() bool {
	if b.predicateOverride != nil {
		return b.predicateOverride()
	}

	return len(b.queryParams[b.requestParamName]) != 0
}

// getPassable return known passable for DatabaseOperator's
func (b *DatabaseOperator) getPassable(res pipeline.Result) (*database.DB, error) {
	casted, ok := res.GetPassable().(*database.DB)
	if !ok {
		return nil, errors.New("invalid result passable %s", reflect.TypeOf(res.GetPassable()).String())
	}

	return casted, nil
}
