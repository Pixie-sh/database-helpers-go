package database

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pixie-sh/errors-go"

	databaserrors "github.com/pixie-sh/database-helpers-go/errors"
)

const (
	// DefaultElasticTimeout is applied when ElasticConfiguration.Timeout <= 0.
	DefaultElasticTimeout = 5 * time.Second
	// DefaultElasticPerPage is the default page size used by ElasticPerPage
	// when no per_page query param is provided.
	DefaultElasticPerPage = 30
	// MaxElasticPerPage is the conservative upper bound for ElasticPerPage.
	MaxElasticPerPage = 100

	contentTypeJSON   = "application/json"
	contentTypeNDJSON = "application/x-ndjson"
)

// ElasticConfiguration holds the connection settings shared by every
// repository talking to a given Elasticsearch cluster.
type ElasticConfiguration struct {
	BaseURL  string        `json:"base_url" toml:"base_url" mapstructure:"base_url"`
	Index    string        `json:"index" toml:"index" mapstructure:"index"`
	Username string        `json:"username" toml:"username" mapstructure:"username"`
	Password string        `json:"password" toml:"password" mapstructure:"password"`
	Timeout  time.Duration `json:"timeout" toml:"timeout" mapstructure:"timeout"`
}

// ElasticSearchRequest is the canonical search request envelope produced by
// the elastic builder and consumed by ElasticRepository.Search.
// Re-exported as elastic.SearchRequest from pipeline/operators/elastic.
type ElasticSearchRequest struct {
	Method string                 `json:"method"`
	Path   string                 `json:"path"`
	Index  string                 `json:"index"`
	Body   map[string]interface{} `json:"body"`
}

// ElasticSearchResponse is the canonical shape returned by the
// Elasticsearch _search endpoint. Only fields commonly consumed are decoded.
type ElasticSearchResponse struct {
	Hits ElasticSearchHits `json:"hits"`
}

// ElasticSearchHits wraps the hits envelope. Total is left as interface{}
// because Elasticsearch returns either a number (pre-7.0) or an object with
// a value field (post-7.0). Use ElasticTotalHits to normalise it.
type ElasticSearchHits struct {
	Total interface{}        `json:"total"`
	Hits  []ElasticSearchHit `json:"hits"`
}

// ElasticSearchHit is a single hit in a search response.
type ElasticSearchHit struct {
	Index  string                 `json:"_index"`
	ID     string                 `json:"_id"`
	Score  *float64               `json:"_score,omitempty"`
	Source map[string]interface{} `json:"_source"`
	Sort   []interface{}          `json:"sort,omitempty"`
}

// ElasticBulkResponse is the shape returned by the Elasticsearch _bulk endpoint.
type ElasticBulkResponse struct {
	Errors bool              `json:"errors"`
	Items  []ElasticBulkItem `json:"items"`
}

// ElasticBulkItem maps the operation name ("index", "create", "update",
// "delete") to its per-item result. Each item map has a single entry in practice.
type ElasticBulkItem map[string]ElasticBulkItemResult

// ElasticBulkItemResult is the per-item outcome of a bulk operation.
type ElasticBulkItemResult struct {
	Index  string                 `json:"_index"`
	ID     string                 `json:"_id"`
	Status int                    `json:"status"`
	Error  map[string]interface{} `json:"error,omitempty"`
}

// ElasticErrorResponse is the standard error payload returned by Elasticsearch.
type ElasticErrorResponse struct {
	Error  ElasticErrorBody `json:"error"`
	Status int              `json:"status"`
}

// ElasticErrorBody is the inner error body of an ElasticErrorResponse.
type ElasticErrorBody struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

// ElasticBulkDocument is a single document to send through the _bulk endpoint.
// ID is used as the document _id; Body is JSON-marshalled as the document source.
type ElasticBulkDocument struct {
	ID   string
	Body interface{}
}

// ElasticRepository is a generic, composition-friendly Elasticsearch repository.
//
// Specific repositories embed it to reuse the HTTP plumbing, bulk indexing,
// index creation and search execution. Domain-specific concerns (hit decoding,
// index mapping definitions, ID extraction, result wrapping with typed IDs)
// stay in the specific repository.
//
//	type DealsElasticRepository struct {
//	    database.ElasticRepository
//	}
//
//	func NewDealsElasticRepository(cfg database.ElasticConfiguration) DealsElasticRepository {
//	    if cfg.Index == "" { cfg.Index = "deals" }
//	    return DealsElasticRepository{ElasticRepository: database.NewElasticRepository(cfg)}
//	}
type ElasticRepository struct {
	config ElasticConfiguration
	client *http.Client
}

// NewElasticRepository builds an ElasticRepository with the default HTTP client.
// Timeout falls back to DefaultElasticTimeout when not configured.
func NewElasticRepository(config ElasticConfiguration) ElasticRepository {
	if config.Timeout <= 0 {
		config.Timeout = DefaultElasticTimeout
	}
	return ElasticRepository{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// Config returns the configuration this repository was built with.
func (r ElasticRepository) Config() ElasticConfiguration { return r.config }

// Client returns the underlying HTTP client.
func (r ElasticRepository) Client() *http.Client { return r.client }

// Index returns the configured default index.
func (r ElasticRepository) Index() string { return r.config.Index }

// ResolveIndex returns override when non-empty (trimmed), otherwise the
// configured default index.
func (r ElasticRepository) ResolveIndex(override string) string {
	override = strings.TrimSpace(override)
	if override == "" {
		return r.config.Index
	}
	return override
}

// Do executes an arbitrary HTTP request against the configured BaseURL+path.
// It returns the raw response status and body. Errors are returned only for
// transport or configuration failures; HTTP non-2xx responses are surfaced
// as (statusCode, body, nil) so callers can interpret them.
func (r ElasticRepository) Do(ctx context.Context, method, path, contentType string, body []byte) (int, []byte, error) {
	if strings.TrimSpace(r.config.BaseURL) == "" {
		return 0, nil, errors.New("elasticsearch base_url is not configured", errors.FieldError{
			Field:   "base_url",
			Rule:    "notConfigured",
			Message: "elasticsearch base_url is not configured",
		}, databaserrors.ElasticBaseURLNotConfiguredErrorCode)
	}

	u, err := url.Parse(strings.TrimRight(r.config.BaseURL, "/"))
	if err != nil {
		return 0, nil, errors.Wrap(err, "invalid elasticsearch base_url", errors.FieldError{
			Field:   "base_url",
			Rule:    "invalid",
			Param:   r.config.BaseURL,
			Message: err.Error(),
		}, databaserrors.ElasticInvalidBaseURLErrorCode)
	}
	u.Path = strings.TrimRight(u.Path, "/") + path

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return 0, nil, errors.Wrap(err, "create elasticsearch request", errors.FieldError{
			Field:   "request",
			Rule:    "creationFailed",
			Param:   method + " " + u.String(),
			Message: err.Error(),
		}, databaserrors.ElasticRequestCreationErrorCode)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if r.config.Username != "" || r.config.Password != "" {
		req.SetBasicAuth(r.config.Username, r.config.Password)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return 0, nil, errors.Wrap(err, "execute elasticsearch request", errors.FieldError{
			Field:   "request",
			Rule:    "executionFailed",
			Param:   method + " " + u.String(),
			Message: err.Error(),
		}, databaserrors.ElasticRequestExecutionErrorCode)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, errors.Wrap(err, "read elasticsearch response", errors.FieldError{
			Field:   "response.body",
			Rule:    "readFailed",
			Message: err.Error(),
		}, databaserrors.ElasticResponseReadErrorCode)
	}
	return resp.StatusCode, respBody, nil
}

// Search executes the given request and decodes the response into
// ElasticSearchResponse. Use elastic.Builder.SearchRequest() to produce
// the request.
func (r ElasticRepository) Search(ctx context.Context, request ElasticSearchRequest) (ElasticSearchResponse, error) {
	body, err := json.Marshal(request.Body)
	if err != nil {
		return ElasticSearchResponse{}, errors.Wrap(err, "marshal elasticsearch request body", errors.FieldError{
			Field:   "request.body",
			Rule:    "marshalFailed",
			Message: err.Error(),
		}, databaserrors.ElasticEncodingErrorCode)
	}

	status, respBody, err := r.Do(ctx, request.Method, request.Path, contentTypeJSON, body)
	if err != nil {
		return ElasticSearchResponse{}, err
	}
	if !statusOK(status) {
		return ElasticSearchResponse{}, errors.New("elasticsearch search failed with status %d: %s", status, string(respBody), errors.FieldError{
			Field:   "response.status",
			Rule:    "nonOK",
			Param:   strconv.Itoa(status),
			Message: string(respBody),
		}, databaserrors.ElasticSearchFailedErrorCode)
	}

	var response ElasticSearchResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return ElasticSearchResponse{}, errors.Wrap(err, "decode elasticsearch response", errors.FieldError{
			Field:   "response.body",
			Rule:    "unmarshalFailed",
			Message: err.Error(),
		}, databaserrors.ElasticDecodingErrorCode)
	}
	return response, nil
}

// EnsureIndex creates the index with the given mapping. An empty index falls
// back to the configured default. A "resource_already_exists_exception" reply
// is treated as success.
func (r ElasticRepository) EnsureIndex(ctx context.Context, index string, mapping map[string]interface{}) error {
	index = r.ResolveIndex(index)
	if index == "" {
		return errors.New("elasticsearch index is not configured", errors.FieldError{
			Field:   "index",
			Rule:    "notConfigured",
			Message: "elasticsearch index is not configured",
		}, databaserrors.ElasticIndexNotConfiguredErrorCode)
	}

	body, err := json.Marshal(mapping)
	if err != nil {
		return errors.Wrap(err, "marshal elasticsearch index mapping", errors.FieldError{
			Field:   "mapping",
			Rule:    "marshalFailed",
			Param:   index,
			Message: err.Error(),
		}, databaserrors.ElasticEncodingErrorCode)
	}

	status, respBody, err := r.Do(ctx, http.MethodPut, "/"+url.PathEscape(index), contentTypeJSON, body)
	if err != nil {
		return err
	}
	if statusOK(status) {
		return nil
	}

	var errResp ElasticErrorResponse
	if jsonErr := json.Unmarshal(respBody, &errResp); jsonErr == nil && errResp.Error.Type == "resource_already_exists_exception" {
		return nil
	}
	return errors.New("elasticsearch index creation failed with status %d: %s", status, string(respBody), errors.FieldError{
		Field:   "index",
		Rule:    "creationFailed",
		Param:   index,
		Message: string(respBody),
	}, databaserrors.ElasticIndexCreationFailedErrorCode)
}

// BulkIndex indexes the given documents through the _bulk endpoint. An empty
// documents slice is a no-op. An empty index falls back to the configured
// default. Any per-item failure in the bulk response is returned as an error.
func (r ElasticRepository) BulkIndex(ctx context.Context, index string, documents []ElasticBulkDocument) error {
	if len(documents) == 0 {
		return nil
	}

	index = r.ResolveIndex(index)
	if index == "" {
		return errors.New("elasticsearch index is not configured", errors.FieldError{
			Field:   "index",
			Rule:    "notConfigured",
			Message: "elasticsearch index is not configured",
		}, databaserrors.ElasticIndexNotConfiguredErrorCode)
	}

	body, err := buildBulkBody(index, documents)
	if err != nil {
		return err
	}

	status, respBody, err := r.Do(ctx, http.MethodPost, "/_bulk", contentTypeNDJSON, body)
	if err != nil {
		return err
	}
	if !statusOK(status) {
		return errors.New("elasticsearch bulk request failed with status %d: %s", status, string(respBody), errors.FieldError{
			Field:   "response.status",
			Rule:    "nonOK",
			Param:   strconv.Itoa(status),
			Message: string(respBody),
		}, databaserrors.ElasticBulkFailedErrorCode)
	}

	var response ElasticBulkResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return errors.Wrap(err, "decode elasticsearch bulk response", errors.FieldError{
			Field:   "response.body",
			Rule:    "unmarshalFailed",
			Message: err.Error(),
		}, databaserrors.ElasticDecodingErrorCode)
	}
	if failureField := firstBulkFailureField(response.Items); response.Errors || failureField != nil {
		if failureField == nil {
			failureField = &errors.FieldError{
				Field:   "bulk.items",
				Rule:    "itemFailure",
				Message: "bulk response reported errors without item details",
			}
		}
		return errors.New("elasticsearch bulk indexing failed: %s", failureField.Message, failureField, databaserrors.ElasticBulkFailedErrorCode)
	}
	return nil
}

func buildBulkBody(index string, documents []ElasticBulkDocument) ([]byte, error) {
	var body bytes.Buffer
	for _, document := range documents {
		action := map[string]map[string]string{
			"index": {
				"_index": index,
				"_id":    document.ID,
			},
		}
		actionBody, err := json.Marshal(action)
		if err != nil {
			return nil, errors.Wrap(err, "marshal elasticsearch bulk action", errors.FieldError{
				Field:   "bulk.action",
				Rule:    "marshalFailed",
				Param:   document.ID,
				Message: err.Error(),
			}, databaserrors.ElasticEncodingErrorCode)
		}
		documentBody, err := json.Marshal(document.Body)
		if err != nil {
			return nil, errors.Wrap(err, "marshal elasticsearch bulk document", errors.FieldError{
				Field:   "bulk.document",
				Rule:    "marshalFailed",
				Param:   document.ID,
				Message: err.Error(),
			}, databaserrors.ElasticEncodingErrorCode)
		}
		body.Write(actionBody)
		body.WriteByte('\n')
		body.Write(documentBody)
		body.WriteByte('\n')
	}
	return body.Bytes(), nil
}

// ElasticBulkFailureMessage returns a human readable description of the first
// failed bulk item, or an empty string if none failed.
func ElasticBulkFailureMessage(items []ElasticBulkItem) string {
	for _, item := range items {
		for operation, result := range item {
			if result.Error != nil || result.Status >= http.StatusMultipleChoices {
				return fmt.Sprintf("operation=%s index=%s id=%s status=%d error=%v",
					operation, result.Index, result.ID, result.Status, result.Error)
			}
		}
	}
	return ""
}

// firstBulkFailureField returns the first failed bulk item as a structured
// FieldError, or nil if none failed.
func firstBulkFailureField(items []ElasticBulkItem) *errors.FieldError {
	for _, item := range items {
		for operation, result := range item {
			if result.Error != nil || result.Status >= http.StatusMultipleChoices {
				return &errors.FieldError{
					Field: "bulk.items[" + result.ID + "]",
					Rule:  "itemFailure",
					Param: strconv.Itoa(result.Status),
					Message: fmt.Sprintf("operation=%s index=%s id=%s status=%d error=%v",
						operation, result.Index, result.ID, result.Status, result.Error),
				}
			}
		}
	}
	return nil
}

// ElasticPerPage parses the per_page query param, falling back to
// defaultPerPage and clamped by maxPerPage. Pass maxPerPage <= 0 to disable
// the upper bound.
func ElasticPerPage(queryParams map[string][]string, defaultPerPage, maxPerPage int) int {
	perPage := defaultPerPage
	if values := queryParams["per_page"]; len(values) > 0 {
		if parsed, err := strconv.Atoi(values[0]); err == nil && parsed > 0 {
			perPage = parsed
		}
	}
	if maxPerPage > 0 && perPage > maxPerPage {
		return maxPerPage
	}
	return perPage
}

// ElasticCurrentPage parses the page query param, falling back to 0.
func ElasticCurrentPage(queryParams map[string][]string) int {
	if values := queryParams["page"]; len(values) > 0 {
		if parsed, err := strconv.Atoi(values[0]); err == nil && parsed > 0 {
			return parsed
		}
	}
	return 0
}

// TrimCursorHits trims hits to at most perPage entries and reports whether
// more results were available (i.e. hits had more than perPage entries).
// perPage <= 0 leaves hits untouched.
func TrimCursorHits(hits []ElasticSearchHit, perPage int) ([]ElasticSearchHit, bool) {
	if perPage <= 0 || len(hits) <= perPage {
		return hits, false
	}
	return hits[:perPage], true
}

// EncodeSearchAfterCursor encodes an Elasticsearch sort tuple into a
// URL-safe base64 cursor string.
func EncodeSearchAfterCursor(values []interface{}) (string, error) {
	encoded, err := json.Marshal(values)
	if err != nil {
		return "", errors.Wrap(err, "encode search_after cursor", errors.FieldError{
			Field:   "cursor",
			Rule:    "encodeFailed",
			Message: err.Error(),
		}, databaserrors.ElasticCursorEncodeErrorCode)
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

// DecodeSearchAfterCursor decodes a cursor previously produced by
// EncodeSearchAfterCursor. An empty cursor yields a nil slice and no error.
func DecodeSearchAfterCursor(cursor string) ([]interface{}, error) {
	if strings.TrimSpace(cursor) == "" {
		return nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, errors.Wrap(err, "decode search_after cursor", errors.FieldError{
			Field:   "cursor",
			Rule:    "invalidBase64",
			Param:   cursor,
			Message: err.Error(),
		}, databaserrors.ElasticInvalidCursorErrorCode)
	}
	var values []interface{}
	if err := json.Unmarshal(decoded, &values); err != nil {
		return nil, errors.Wrap(err, "parse search_after cursor", errors.FieldError{
			Field:   "cursor",
			Rule:    "invalidJSON",
			Param:   cursor,
			Message: err.Error(),
		}, databaserrors.ElasticInvalidCursorErrorCode)
	}
	return values, nil
}

// ElasticTotalHits extracts the total hit count from the polymorphic "total"
// field returned by Elasticsearch (object form post-7.0, number form pre-7.0).
func ElasticTotalHits(total interface{}) int64 {
	switch typed := total.(type) {
	case map[string]interface{}:
		return Int64FromInterface(typed["value"])
	case float64:
		return int64(typed)
	case int64:
		return typed
	case int:
		return int64(typed)
	default:
		return 0
	}
}

// Int64FromInterface converts a JSON-decoded numeric value into int64.
func Int64FromInterface(value interface{}) int64 {
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int64:
		return typed
	case int:
		return int64(typed)
	default:
		return 0
	}
}

// FirstStringValue returns the first string-coercible value found in values
// under the given keys, in order. Returns "" if no key resolves to a string
// or fmt.Stringer.
func FirstStringValue(values map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		value, ok := values[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case string:
			return typed
		case fmt.Stringer:
			return typed.String()
		}
	}
	return ""
}

func statusOK(status int) bool {
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}
