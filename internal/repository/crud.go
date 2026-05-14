package repository

import "gorm.io/gorm/clause"

// Create inserts a row.
func (c *RepoConfig[T]) Create(row *T) error {
	return c.Query().Create(row).Error
}

// Save inserts or updates (all fields) according to PK.
func (c *RepoConfig[T]) Save(row *T) error {
	return c.Query().Save(row).Error
}

// Updates applies non-zero fields from vals (struct or map[string]any).
func (c *RepoConfig[T]) Updates(where any, vals any) error {
	tx := c.Query().Where(where)
	return tx.Updates(vals).Error
}

// SelectByID returns one row by primary key column "id".
func (c *RepoConfig[T]) SelectByID(id any) (*T, error) {
	var v T
	if err := c.Query().Where("id = ?", id).Take(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

// DeleteByID deletes rows with id column.
func (c *RepoConfig[T]) DeleteByID(id any) error {
	var zero T
	return c.Query().Where("id = ?", id).Delete(&zero).Error
}

// DeleteWhere deletes using arbitrary conditions (same semantics as Where).
func (c *RepoConfig[T]) DeleteWhere(where any, args ...any) error {
	var zero T
	return c.Query().Where(where, args...).Delete(&zero).Error
}

// FindWhere loads rows matching conditions without pagination.
func (c *RepoConfig[T]) FindWhere(dest *[]T, where any, args ...any) error {
	return c.Query().Where(where, args...).Find(dest).Error
}

// FindPage loads a page plus total matching rows (respects filteredQuery).
func (c *RepoConfig[T]) FindPage(opts *ReadOpts, searchColumns []string) ([]T, int64, error) {
	base := c.filteredQuery(opts, searchColumns)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	off, ps, ord := c.Pagination(opts)
	var out []T
	if err := base.Order(ord).Offset(off).Limit(ps).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// Upsert performs ON CONFLICT — requires PostgreSQL-compatible clause (e.g. on "id").
func (c *RepoConfig[T]) Upsert(row *T, conflictColumns []string, updateColumns []string) error {
	return c.Query().Clauses(clause.OnConflict{
		Columns:   conflictClauseColumns(conflictColumns),
		DoUpdates: clause.AssignmentColumns(updateColumns),
	}).Create(row).Error
}

func conflictClauseColumns(names []string) []clause.Column {
	cols := make([]clause.Column, 0, len(names))
	for _, n := range names {
		if s, ok := sanitizeIdent(n); ok {
			cols = append(cols, clause.Column{Name: s})
		}
	}
	return cols
}
