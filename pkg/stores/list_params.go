package stores

import (
	"fmt"
	"regexp"

	"gorm.io/gorm"
)

var validFieldName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type Filter struct {
	Field string
	Value any
}

type ListParams struct {
	Filters  []Filter
	Page     int
	PageSize int
	OrderBy  string
}

func (p ListParams) Limit() int {
	if p.PageSize <= 0 {
		return DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		return MaxPageSize
	}
	return p.PageSize
}

func (p ListParams) Offset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit()
}

func (p ListParams) WithDefaultOrder(order string) ListParams {
	if p.OrderBy == "" {
		p.OrderBy = order
	}
	return p
}

func (p ListParams) Apply(query *gorm.DB) *gorm.DB {
	for _, f := range p.Filters {
		if validFieldName.MatchString(f.Field) {
			query = query.Where(fmt.Sprintf("%s = ?", f.Field), f.Value)
		}
	}
	if p.OrderBy != "" {
		query = query.Order(p.OrderBy)
	}
	query = query.Limit(p.Limit()).Offset(p.Offset())
	return query
}

func (p ListParams) ApplyFiltersOnly(query *gorm.DB) *gorm.DB {
	for _, f := range p.Filters {
		if validFieldName.MatchString(f.Field) {
			query = query.Where(fmt.Sprintf("%s = ?", f.Field), f.Value)
		}
	}
	return query
}

func (p ListParams) TotalPages(total int64) int {
	limit := p.Limit()
	pages := int(total) / limit
	if int(total)%limit > 0 {
		pages++
	}
	return pages
}
