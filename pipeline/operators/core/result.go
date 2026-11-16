package core

// Result is the value carried through a Pipeline. Concrete operators write
// to (and read from) the passable held inside a Result implementation.
type Result interface {
	WithPassable(passable interface{})
	GetPassable() interface{}
	Error() error
}

// BaseResult is the default Result implementation used by every operator.
// The previous field tracks the operator that produced this result, which
// lets stateful operators (notably the SQL aggregator) influence how the
// next operator combines its clause with the running query.
type BaseResult struct {
	passable interface{}
	error    error

	previous Operator
}

// NewResult builds a BaseResult around the given passable.
func NewResult(passable interface{}) *BaseResult {
	return &BaseResult{
		passable: passable,
	}
}

// GetPassable returns the value currently carried by this result.
func (r *BaseResult) GetPassable() interface{} {
	return r.passable
}

// Error returns any error attached to this result.
func (r *BaseResult) Error() error {
	return r.error
}

// WithPassable replaces the carried passable.
func (r *BaseResult) WithPassable(passable interface{}) {
	r.passable = passable
}

// Previous returns the operator that produced this result, or nil if none
// has been set. Used by SQL operators to detect when a prior aggregator
// operator has requested a specific boolean combinator.
func (r *BaseResult) Previous() Operator {
	return r.previous
}

// SetPrevious records the operator that produced this result.
func (r *BaseResult) SetPrevious(op Operator) {
	r.previous = op
}

// UntypedPaginatedResult is the on-the-wire shape of a page-based paginated
// API response. Data is carried as interface{} so the same struct can be
// embedded by the typed PaginatedResult[T] generic alias.
type UntypedPaginatedResult struct {
	Data             interface{}         `json:"data"`
	PerPage          int                 `json:"per_page"`
	CurrentPage      int                 `json:"current_page"`
	TotalResults     int64               `json:"total_results"`
	PageCount        int64               `json:"page_count"`
	AvailablePerPage []int               `json:"available_per_page"`
	QueryParams      map[string][]string `json:"query_params"`
}

// PaginatedResult is the typed counterpart of UntypedPaginatedResult.
type PaginatedResult[T any] struct {
	UntypedPaginatedResult

	Data T `json:"data"`
}

// UntypedOffsetPaginatedResult is the on-the-wire shape of an offset/limit
// paginated API response.
type UntypedOffsetPaginatedResult struct {
	Data             interface{}         `json:"data"`
	PerPage          int                 `json:"per_page"`
	CurrentPage      int                 `json:"current_page"`
	HasMore          bool                `json:"has_more"`
	AvailablePerPage []int               `json:"available_per_page"`
	QueryParams      map[string][]string `json:"query_params"`
}

// OffsetPaginatedResult is the typed counterpart of UntypedOffsetPaginatedResult.
type OffsetPaginatedResult[T any] struct {
	UntypedOffsetPaginatedResult

	Data T `json:"data"`
}

// UntypedListedResult is the on-the-wire shape of a list (non-paginated)
// API response.
type UntypedListedResult struct {
	Data         interface{}         `json:"data"`
	TotalResults int64               `json:"total_results"`
	QueryParams  map[string][]string `json:"query_params"`
}

// ListedResult is the typed counterpart of UntypedListedResult.
type ListedResult[T any] struct {
	UntypedListedResult

	Data T `json:"data"`
}

// UntypedNextAfterPaginatedResult is the on-the-wire shape of a cursor
// (search_after) paginated API response, used primarily by Elasticsearch.
type UntypedNextAfterPaginatedResult struct {
	Data             interface{}         `json:"data"`
	PerPage          int                 `json:"per_page"`
	CurrentPage      int                 `json:"current_page"`
	TotalResults     int64               `json:"total_results"`
	PageCount        int64               `json:"page_count"`
	NextSearchAfter  *string             `json:"next_search_after,omitempty"`
	HasMore          bool                `json:"has_more,omitempty"`
	AvailablePerPage []int               `json:"available_per_page"`
	QueryParams      map[string][]string `json:"query_params"`
	RandomSeed       *int64              `json:"random_seed,omitempty"`
}

// NextAfterPaginatedResult is the typed counterpart of UntypedNextAfterPaginatedResult.
type NextAfterPaginatedResult[T any] struct {
	UntypedNextAfterPaginatedResult

	Data T `json:"data"`
}
