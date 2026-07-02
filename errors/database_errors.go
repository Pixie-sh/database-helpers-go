package errors

import "github.com/pixie-sh/errors-go"

// DatabaseErrorCode is the numeric base for pipeline operator error
// codes. Each individual code adds an HTTP status suffix so that
// `code % 1000` yields a valid HTTP status, as required by errors-go.
const DatabaseErrorCode = 71000

var (
	// ElasticErrorErrorCode is the generic fallback for elasticsearch failures.
	// Prefer the concrete codes below.
	ElasticErrorErrorCode = errors.NewErrorCode("ElasticErrorErrorCode", DatabaseErrorCode+errors.HTTPServerError)

	// Configuration errors — server misconfigured.
	ElasticBaseURLNotConfiguredErrorCode = errors.NewErrorCode("ElasticBaseURLNotConfiguredErrorCode", DatabaseErrorCode+errors.HTTPServerError)
	ElasticInvalidBaseURLErrorCode       = errors.NewErrorCode("ElasticInvalidBaseURLErrorCode", DatabaseErrorCode+errors.HTTPServerError)
	ElasticIndexNotConfiguredErrorCode   = errors.NewErrorCode("ElasticIndexNotConfiguredErrorCode", DatabaseErrorCode+errors.HTTPServerError)

	// HTTP transport — upstream/network failures.
	ElasticRequestCreationErrorCode  = errors.NewErrorCode("ElasticRequestCreationErrorCode", DatabaseErrorCode+errors.HTTPServerError)
	ElasticRequestExecutionErrorCode = errors.NewErrorCode("ElasticRequestExecutionErrorCode", DatabaseErrorCode+errors.HTTPServerError)
	ElasticResponseReadErrorCode     = errors.NewErrorCode("ElasticResponseReadErrorCode", DatabaseErrorCode+errors.HTTPServerError)

	// Serialization — request/response encoding failures.
	ElasticEncodingErrorCode = errors.NewErrorCode("ElasticEncodingErrorCode", DatabaseErrorCode+errors.HTTPServerError)
	ElasticDecodingErrorCode = errors.NewErrorCode("ElasticDecodingErrorCode", DatabaseErrorCode+errors.HTTPServerError)

	// Elasticsearch API non-2xx responses.
	ElasticSearchFailedErrorCode        = errors.NewErrorCode("ElasticSearchFailedErrorCode", DatabaseErrorCode+errors.HTTPServerError)
	ElasticIndexCreationFailedErrorCode = errors.NewErrorCode("ElasticIndexCreationFailedErrorCode", DatabaseErrorCode+errors.HTTPServerError)
	ElasticBulkFailedErrorCode          = errors.NewErrorCode("ElasticBulkFailedErrorCode", DatabaseErrorCode+errors.HTTPServerError)

	// Pagination cursor errors.
	ElasticCursorEncodeErrorCode  = errors.NewErrorCode("ElasticCursorEncodeErrorCode", DatabaseErrorCode+errors.HTTPServerError)
	ElasticInvalidCursorErrorCode = errors.NewErrorCode("ElasticInvalidCursorErrorCode", DatabaseErrorCode+errors.HTTPInvalidData)
)
