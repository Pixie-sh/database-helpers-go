package elastic

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/pixie-sh/database-helpers-go/pipeline/operators/operator_errors"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"
)

type Clause string

const (
	ClauseFilter  Clause = "filter"
	ClauseMust    Clause = "must"
	ClauseMustNot Clause = "must_not"
	ClauseShould  Clause = "should"
)

type SourceOperator struct {
	ElasticOperator
	fields []string
}

func NewSourceOperator(fields ...string) *SourceOperator {
	return &SourceOperator{fields: fields}
}

func (op *SourceOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	builder, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(operator_errors.InvalidPassableErrorCode)
	}

	builder.SetSource(op.fields...)
	genericResult.WithPassable(builder)
	return genericResult, nil
}

type QueryOperator struct {
	ElasticOperator
	clause Query
}

func NewQueryOperator(query Query) *QueryOperator {
	return &QueryOperator{clause: query}
}

func (op *QueryOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	builder, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(operator_errors.InvalidPassableErrorCode)
	}

	builder.SetQuery(op.clause)
	genericResult.WithPassable(builder)
	return genericResult, nil
}

type BoolClauseOperator struct {
	ElasticOperator
	clause Clause
	query  Query
}

func NewBoolClauseOperator(clause Clause, query Query) *BoolClauseOperator {
	return &BoolClauseOperator{clause: clause, query: query}
}

func NewFilterOperator(query Query) *BoolClauseOperator {
	return NewBoolClauseOperator(ClauseFilter, query)
}

func NewMustOperator(query Query) *BoolClauseOperator {
	return NewBoolClauseOperator(ClauseMust, query)
}

func NewMustNotOperator(query Query) *BoolClauseOperator {
	return NewBoolClauseOperator(ClauseMustNot, query)
}

func NewShouldOperator(query Query) *BoolClauseOperator {
	return NewBoolClauseOperator(ClauseShould, query)
}

func (op *BoolClauseOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	builder, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(operator_errors.InvalidPassableErrorCode)
	}

	switch op.clause {
	case ClauseFilter:
		builder.AddFilter(op.query)
	case ClauseMust:
		builder.AddMust(op.query)
	case ClauseMustNot:
		builder.AddMustNot(op.query)
	case ClauseShould:
		builder.AddShould(op.query)
	default:
		return nil, errors.New("unknown elastic bool clause %s", op.clause).WithErrorCode(operator_errors.UnknownElasticBoolClauseErrorCode)
	}

	genericResult.WithPassable(builder)
	return genericResult, nil
}

type MinimumShouldMatchOperator struct {
	ElasticOperator
	value int
}

func NewMinimumShouldMatchOperator(value int) *MinimumShouldMatchOperator {
	return &MinimumShouldMatchOperator{value: value}
}

func (op *MinimumShouldMatchOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	builder, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(operator_errors.InvalidPassableErrorCode)
	}

	builder.SetMinimumShouldMatch(op.value)
	genericResult.WithPassable(builder)
	return genericResult, nil
}

type RawBodyOperator struct {
	ElasticOperator
	body map[string]interface{}
}

func NewRawBodyOperator(body map[string]interface{}) *RawBodyOperator {
	return &RawBodyOperator{body: body}
}

func (op *RawBodyOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	builder, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(operator_errors.InvalidPassableErrorCode)
	}

	for key, value := range op.body {
		builder.Body()[key] = value
	}
	genericResult.WithPassable(builder)
	return genericResult, nil
}

type ScriptFragmentOperator struct {
	ElasticOperator
	section ScriptSection
	source  string
	params  map[string]interface{}
}

func NewScriptFragmentOperator(section ScriptSection, source string, params map[string]interface{}) *ScriptFragmentOperator {
	return &ScriptFragmentOperator{section: section, source: source, params: params}
}

func NewScriptHelperOperator(source string, params map[string]interface{}) *ScriptFragmentOperator {
	return NewScriptFragmentOperator(ScriptSectionHelpers, source, params)
}

func NewScriptSetupOperator(source string, params map[string]interface{}) *ScriptFragmentOperator {
	return NewScriptFragmentOperator(ScriptSectionSetup, source, params)
}

func NewScriptScoringOperator(source string, params map[string]interface{}) *ScriptFragmentOperator {
	return NewScriptFragmentOperator(ScriptSectionScoring, source, params)
}

func NewScriptReturnOperator(source string, params map[string]interface{}) *ScriptFragmentOperator {
	return NewScriptFragmentOperator(ScriptSectionReturn, source, params)
}

func (op *ScriptFragmentOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	builder, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(operator_errors.InvalidPassableErrorCode)
	}

	script := builder.ScriptBuilder()
	script.Add(op.section, op.source)
	if err := script.AddParams(op.params); err != nil {
		return nil, err
	}
	genericResult.WithPassable(builder)
	return genericResult, nil
}

type ScriptScoreOperator struct {
	ElasticOperator
	query Query
}

func NewScriptScoreOperator(query Query) *ScriptScoreOperator {
	return &ScriptScoreOperator{query: query}
}

func (op *ScriptScoreOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	builder, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(operator_errors.InvalidPassableErrorCode)
	}

	if err := builder.ApplyScriptScore(op.query, builder.ScriptBuilder()); err != nil {
		return nil, err
	}

	genericResult.WithPassable(builder)
	return genericResult, nil
}

type DebugOperator struct {
	ElasticOperator
	log          logger.Interface
	pretty       bool
	includeIndex bool
	maxBodyBytes int
	redactParams []string
}

func NewDebugOperator(log logger.Interface, options ...DebugOption) *DebugOperator {
	op := &DebugOperator{log: log, pretty: true, includeIndex: true}
	for _, option := range options {
		option(op)
	}
	return op
}

type DebugOption func(*DebugOperator)

func DebugCompact() DebugOption {
	return func(op *DebugOperator) { op.pretty = false }
}

func DebugMaxBodyBytes(max int) DebugOption {
	return func(op *DebugOperator) { op.maxBodyBytes = max }
}

func DebugRedactParams(keys ...string) DebugOption {
	return func(op *DebugOperator) { op.redactParams = keys }
}

func DebugWithoutIndex() DebugOption {
	return func(op *DebugOperator) { op.includeIndex = false }
}

func (op *DebugOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	builder, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(operator_errors.InvalidPassableErrorCode)
	}

	body := cloneMap(builder.Body())
	redactScriptParams(body, op.redactParams)
	encoded, err := marshalBody(body, op.pretty)
	if err != nil {
		return nil, err
	}

	message := string(encoded)
	if op.maxBodyBytes > 0 && len(message) > op.maxBodyBytes {
		message = message[:op.maxBodyBytes] + "..."
	}
	if op.includeIndex {
		message = "POST /" + strings.Trim(builder.Index(), "/") + "/_search\n" + message
	}

	if op.log != nil {
		op.log.With("elastic_request", message).Debug("elastic search request")
	}

	genericResult.WithPassable(builder)
	return genericResult, nil
}

func marshalBody(body map[string]interface{}, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(body, "", "  ")
	}
	return json.Marshal(body)
}

func cloneMap(source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(source))
	for key, value := range source {
		result[key] = cloneValue(value)
	}
	return result
}

func cloneValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return cloneMap(typed)
	case Query:
		return cloneMap(map[string]interface{}(typed))
	case []interface{}:
		result := make([]interface{}, len(typed))
		for i, item := range typed {
			result[i] = cloneValue(item)
		}
		return result
	default:
		return value
	}
}

func redactScriptParams(body map[string]interface{}, keys []string) {
	if len(keys) == 0 {
		return
	}

	redact := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		redact[key] = struct{}{}
	}
	walkRedact(body, redact)
}

func walkRedact(value interface{}, redact map[string]struct{}) {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, item := range typed {
			if key == "params" {
				if params, ok := item.(map[string]interface{}); ok {
					for param := range redact {
						if _, exists := params[param]; exists {
							params[param] = "[REDACTED]"
						}
					}
				}
			}
			walkRedact(item, redact)
		}
	case []interface{}:
		for _, item := range typed {
			walkRedact(item, redact)
		}
	}
}
