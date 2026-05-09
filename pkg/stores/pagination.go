package stores

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PaginationParams struct {
	Page     int
	PageSize int
}

func (p PaginationParams) Offset() int {
	if p.Page <= 0 {
		return 0
	}
	return (p.Page - 1) * p.Limit()
}

func (p PaginationParams) Limit() int {
	if p.PageSize <= 0 {
		return DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		return MaxPageSize
	}
	return p.PageSize
}

type PaginatedResult[T any] struct {
	Items      []T
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}
