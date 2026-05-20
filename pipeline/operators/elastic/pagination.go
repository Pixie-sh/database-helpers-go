package elastic

import (
	"context"
	"strconv"

	"github.com/pixie-sh/errors-go"
)

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
		return nil, errors.NewWithError(err, "invalid passable")
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
		return nil, errors.NewWithError(err, "invalid passable")
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
		return nil, errors.NewWithError(err, "invalid passable")
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
		return nil, errors.NewWithError(err, "invalid passable")
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
		return nil, errors.NewWithError(err, "invalid passable")
	}

	builder.SetTrackTotalHits(op.value)
	genericResult.WithPassable(builder)
	return genericResult, nil
}
