package operator_errors

import (
	"github.com/pixie-sh/errors-go"
)

// PipelineOperatorErrorCode is the numeric base for pipeline operator error
// codes. Each individual code adds an HTTP status suffix so that
// `code % 1000` yields a valid HTTP status, as required by errors-go.
const PipelineOperatorErrorCode = 70000

var (
	InvalidPassableErrorCode       = errors.NewErrorCode("InvalidPassableErrorCode", PipelineOperatorErrorCode+errors.HTTPServerError)
	InvalidResultPassableErrorCode = errors.NewErrorCode("InvalidResultPassableErrorCode", PipelineOperatorErrorCode+errors.HTTPServerError)
	InvalidResultTypeErrorCode     = errors.NewErrorCode("InvalidResultTypeErrorCode", PipelineOperatorErrorCode+errors.HTTPServerError)

	IdsLimitExceededErrorCode = errors.NewErrorCode("IdsLimitExceededErrorCode", PipelineOperatorErrorCode+errors.HTTPInvalidData)
	InvalidULIDErrorCode      = errors.NewErrorCode("InvalidULIDErrorCode", PipelineOperatorErrorCode+errors.HTTPInvalidData)
	EmptyULIDErrorCode        = errors.NewErrorCode("EmptyULIDErrorCode", PipelineOperatorErrorCode+errors.HTTPInvalidData)

	UnknownElasticBoolClauseErrorCode = errors.NewErrorCode("UnknownElasticBoolClauseErrorCode", PipelineOperatorErrorCode+errors.HTTPBadRequest)
	NilScriptBuilderErrorCode         = errors.NewErrorCode("NilScriptBuilderErrorCode", PipelineOperatorErrorCode+errors.HTTPServerError)
	EmptyScriptSourceErrorCode        = errors.NewErrorCode("EmptyScriptSourceErrorCode", PipelineOperatorErrorCode+errors.HTTPInvalidData)
	DuplicateScriptParamErrorCode     = errors.NewErrorCode("DuplicateScriptParamErrorCode", PipelineOperatorErrorCode+errors.HTTPInvalidData)
)
