package elastic

import (
	"context"
	"reflect"

	base "github.com/pixie-sh/database-helpers-go/pipeline/operators"
	"github.com/pixie-sh/errors-go"
)

type Query map[string]interface{}

type SearchRequest struct {
	Method string                 `json:"method"`
	Path   string                 `json:"path"`
	Index  string                 `json:"index"`
	Body   map[string]interface{} `json:"body"`
}

type Result = base.Result

type Operator = base.Operator

type QueryParams = base.QueryParams

type ElasticOperator struct {
	predicateOverride func() bool
	queryParams       QueryParams
	requestParamName  string
}

func (op *ElasticOperator) GetType() uint {
	return base.ElasticOperatorType
}

func (op *ElasticOperator) Predicate(_ context.Context, ignoreOverride bool) bool {
	if op.predicateOverride != nil && !ignoreOverride {
		return op.predicateOverride()
	}
	if op.requestParamName == "" {
		return true
	}

	return len(op.queryParams[op.requestParamName]) != 0
}

func (op *ElasticOperator) getPassable(res Result) (*Builder, error) {
	casted, ok := res.GetPassable().(*Builder)
	if !ok {
		return nil, errors.New("invalid result passable %s", reflect.TypeOf(res.GetPassable()).String())
	}

	return casted, nil
}
