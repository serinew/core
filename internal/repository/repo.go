package repository

import (
	"regexp"
	"strings"
	"sync"

	"gorm.io/gorm"
)

var identRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// EscapeLike escapes LIKE / ILIKE metacharacters (\, %, _).
func EscapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// RepoConfig is a typed table-scope helper bound to Store.
type RepoConfig[T any] struct {
	store *Store
	Table string
}

// Repo builds a lazily initialised closure (once per closure instance).
// Use Repo[MyRow](store, "my_table").
func Repo[T any](s *Store, table string) func() *RepoConfig[T] {
	var r *RepoConfig[T]
	var once sync.Once
	return func() *RepoConfig[T] {
		once.Do(func() {
			r = &RepoConfig[T]{store: s, Table: strings.TrimSpace(table)}
		})
		return r
	}
}

// Query returns DB scoped to Table() for chaining.
func (c *RepoConfig[T]) Query() *gorm.DB {
	return c.store.DB.Table(c.Table)
}

func sanitizeIdent(col string) (string, bool) {
	col = strings.TrimSpace(col)
	if col == "" || !identRegex.MatchString(col) {
		return "", false
	}
	return col, true
}

func pickSearchColumns(opts *ReadOpts, fallback []string) []string {
	if opts != nil && len(opts.Columns) > 0 {
		out := make([]string, 0, len(opts.Columns))
		for _, col := range opts.Columns {
			if s, ok := sanitizeIdent(col); ok {
				out = append(out, s)
			}
		}
		return out
	}
	out := make([]string, 0, len(fallback))
	for _, col := range fallback {
		if s, ok := sanitizeIdent(col); ok {
			out = append(out, s)
		}
	}
	return out
}

func sanitizeOrder(opts *ReadOpts, defaultOrder string) string {
	if opts != nil && opts.OrderBy != nil {
		s := strings.TrimSpace(*opts.OrderBy)
		if s != "" {
			parts := strings.Split(s, ",")
			var b strings.Builder
			first := true
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				fields := strings.Fields(p)
				if len(fields) == 0 {
					continue
				}
				col, ok := sanitizeIdent(fields[0])
				if !ok {
					continue
				}
				dir := "ASC"
				if len(fields) > 1 {
					switch strings.ToUpper(fields[1]) {
					case "ASC", "DESC":
						dir = strings.ToUpper(fields[1])
					default:
						continue
					}
				}
				if !first {
					b.WriteString(", ")
				}
				first = false
				b.WriteString(`"`)
				b.WriteString(col)
				b.WriteString(`" `)
				b.WriteString(dir)
			}
			if b.Len() > 0 {
				return b.String()
			}
		}
	}
	return defaultOrder
}

// filteredQuery applies search filters (PostgreSQL ILIKE) without ORDER / LIMIT.
func (c *RepoConfig[T]) filteredQuery(opts *ReadOpts, searchColumns []string) *gorm.DB {
	q := c.Query()

	search := ""
	if opts != nil && opts.Search != nil {
		search = strings.TrimSpace(*opts.Search)
	}
	cols := pickSearchColumns(opts, searchColumns)
	if search != "" && len(cols) > 0 {
		pattern := "%" + EscapeLike(search) + "%"
		var ors []string
		var args []any
		for _, col := range cols {
			ors = append(ors, `CAST("`+col+`" AS text) ILIKE ? ESCAPE '\'`)
			args = append(args, pattern)
		}
		q = q.Where(strings.Join(ors, " OR "), args...)
	}

	return q
}

// Pagination returns offset, limit, ORDER BY clause fragment.
func (c *RepoConfig[T]) Pagination(opts *ReadOpts) (offset, pageSize int, orderClause string) {
	page, pageSize := 1, 20

	if opts != nil {
		if opts.Page != nil && *opts.Page >= 1 {
			page = *opts.Page
		}
		if opts.PageSize != nil {
			pageSize = *opts.PageSize
			if pageSize > 100 {
				pageSize = 100
			}
			if pageSize < 1 {
				pageSize = 20
			}
		}
	}

	orderClause = sanitizeOrder(opts, `"id" ASC`)
	offset = (page - 1) * pageSize
	return offset, pageSize, orderClause
}

// ListQuery returns a chained query with filters + order (no OFFSET/LIMIT).
func (c *RepoConfig[T]) ListQuery(opts *ReadOpts, searchColumns []string) (query *gorm.DB, offset, pageSize int) {
	q := c.filteredQuery(opts, searchColumns)
	off, ps, ord := c.Pagination(opts)
	return q.Order(ord), off, ps
}
