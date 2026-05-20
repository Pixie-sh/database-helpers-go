package elastic

import (
	"context"
	"math"
	"strconv"

	"github.com/pixie-sh/database-helpers-go/database"
	databaserrors "github.com/pixie-sh/database-helpers-go/errors"
	base "github.com/pixie-sh/database-helpers-go/pipeline/operators"
	"github.com/pixie-sh/errors-go"
)

// BuildPaginatedResult constructs page-based pagination metadata from the
// request query params and the total number of hits returned by Elasticsearch.
//
// paginationOptions advertises the per_page values exposed to clients; the
// first entry is the default per_page, the last entry is the upper bound.
// When paginationOptions is empty, database.DefaultElasticPerPage is used and
// no upper bound is enforced.
func BuildPaginatedResult(queryParams base.QueryParams, data interface{}, totalResults int64, paginationOptions ...int) base.UntypedPaginatedResult {
	perPage := database.DefaultElasticPerPage
	if len(paginationOptions) > 0 {
		perPage = paginationOptions[0]
	}
	if values := queryParams["per_page"]; len(values) > 0 {
		if parsed, err := strconv.Atoi(values[0]); err == nil && parsed > 0 {
			perPage = parsed
		}
	}
	if len(paginationOptions) > 0 {
		if max := paginationOptions[len(paginationOptions)-1]; max > 0 && perPage > max {
			perPage = max
		}
	}

	currentPage := database.ElasticCurrentPage(queryParams)

	pageCount := int64(0)
	if totalResults > 0 && perPage > 0 {
		pageCount = int64(math.Ceil(float64(totalResults) / float64(perPage)))
	}

	return base.UntypedPaginatedResult{
		Data:             data,
		PerPage:          perPage,
		CurrentPage:      currentPage,
		TotalResults:     totalResults,
		PageCount:        pageCount,
		AvailablePerPage: paginationOptions,
		QueryParams:      queryParams,
	}
}

type FromSizePaginationOperator struct {
	ElasticOperator
	from int
	size int
}

func NewFromSizePaginationOperator(from int, size int) *FromSizePaginationOperator {
	return &FromSizePaginationOperator{from: from, size: size}
}

func NewOffsetPaginationOperator(queryParams QueryParams, defaultSize int, maxSize int) *FromSizePaginationOperator {
	size := defaultSize
	if values := queryParams["per_page"]; len(values) > 0 {
		if parsed, err := strconv.Atoi(values[0]); err == nil && parsed > 0 {
			size = parsed
		}
	}
	if maxSize > 0 && size > maxSize {
		size = maxSize
	}

	page := 0
	if values := queryParams["page"]; len(values) > 0 {
		if parsed, err := strconv.Atoi(values[0]); err == nil && parsed > 0 {
			page = parsed
		}
	}

	return NewFromSizePaginationOperator(page*size, size)
}

func (op *FromSizePaginationOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	builder, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(databaserrors.InvalidPassableErrorCode)
	}

	builder.SetFrom(op.from).SetSize(op.size)
	genericResult.WithPassable(builder)
	return genericResult, nil
}

type SearchAfterPaginationOperator struct {
	ElasticOperator
	size        int
	sort        []Query
	searchAfter []interface{}
}

func NewSearchAfterPaginationOperator(size int, sort []Query, searchAfter ...interface{}) *SearchAfterPaginationOperator {
	return &SearchAfterPaginationOperator{size: size, sort: sort, searchAfter: searchAfter}
}

func (op *SearchAfterPaginationOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	builder, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(databaserrors.InvalidPassableErrorCode)
	}

	builder.SetSize(op.size)
	if len(op.sort) > 0 {
		builder.SetSort(op.sort...)
	}
	if len(op.searchAfter) > 0 {
		builder.SetSearchAfter(op.searchAfter...)
	}
	genericResult.WithPassable(builder)
	return genericResult, nil
}

type PointInTimeOperator struct {
	ElasticOperator
	id        string
	keepAlive string
}

func NewPointInTimeOperator(id string, keepAlive string) *PointInTimeOperator {
	return &PointInTimeOperator{id: id, keepAlive: keepAlive}
}

func (op *PointInTimeOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	builder, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(databaserrors.InvalidPassableErrorCode)
	}

	builder.SetPointInTime(op.id, op.keepAlive)
	genericResult.WithPassable(builder)
	return genericResult, nil
}

type SortOperator struct {
	ElasticOperator
	sort []Query
}

func NewSortOperator(sort ...Query) *SortOperator {
	return &SortOperator{sort: sort}
}

func (op *SortOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	builder, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(databaserrors.InvalidPassableErrorCode)
	}

	builder.SetSort(op.sort...)
	genericResult.WithPassable(builder)
	return genericResult, nil
}

type TrackTotalHitsOperator struct {
	ElasticOperator
	value interface{}
}

func NewTrackTotalHitsOperator(value interface{}) *TrackTotalHitsOperator {
	return &TrackTotalHitsOperator{value: value}
}

func (op *TrackTotalHitsOperator) Handle(_ context.Context, genericResult Result) (Result, error) {
	builder, err := op.getPassable(genericResult)
	if err != nil {
		return nil, errors.NewWithError(err, "invalid passable").WithErrorCode(databaserrors.InvalidPassableErrorCode)
	}

	builder.SetTrackTotalHits(op.value)
	genericResult.WithPassable(builder)
	return genericResult, nil
}
