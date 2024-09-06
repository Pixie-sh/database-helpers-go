package operators

// Result interface to avoid circular imports
type Result interface {
	WithPassable(passable interface{})
	GetPassable() interface{}
	Error() error
}

// BaseResult something
type BaseResult struct {
	passable interface{}
	error    error

	previous Operator
}

// NewResult something amazing, is it?
func NewResult(passable interface{}) *BaseResult {
	return &BaseResult{
		passable: passable,
	}
}

// GetPassable one more comment for you
func (r *BaseResult) GetPassable() interface{} {
	return r.passable
}

// Error another one
func (r *BaseResult) Error() error {
	return r.error
}

// WithPassable one more
func (r *BaseResult) WithPassable(passable interface{}) {
	r.passable = passable
}

// UntypedPaginatedResult struct received when apis with pagination are called
type UntypedPaginatedResult struct {
	Data             interface{}         `json:"data"`
	PerPage          int                 `json:"per_page"`
	CurrentPage      int                 `json:"current_page"`
	TotalResults     int64               `json:"total_results"`
	PageCount        int64               `json:"page_count"`
	AvailablePerPage []int               `json:"available_per_page"`
	QueryParams      map[string][]string `json:"query_params"`
}

// PaginatedResult struct received when apis with pagination are called
type PaginatedResult[T any] struct {
	UntypedPaginatedResult

	Data T `json:"data"`
}
