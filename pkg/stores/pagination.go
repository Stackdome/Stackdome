package stores

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PaginatedResult[T any] struct {
	Items      []T
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}
