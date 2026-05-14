package repository

import "github.com/serinew/core/internal/types"

// ReadOpts drives list pagination, search (ILIKE), and ordering.
type ReadOpts struct {
	Page     *int
	PageSize *int
	Search   *string
	Columns  []string // when Search set: optional override of handler searchColumns
	OrderBy  *string
}

// ReadOptsFromListQuery maps HTTP query params helper to repository options.
func ReadOptsFromListQuery(l *types.ListQuery) *ReadOpts {
	if l == nil {
		return nil
	}
	return &ReadOpts{
		Page:     l.Page,
		PageSize: l.PageSize,
		Search:   l.Search,
	}
}
