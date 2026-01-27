package goSqlite

import (
	"fmt"
	"strings"
)

func (b *Builder) OrWhereEq(column string, value any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.OrWhere(fmt.Sprintf("%s = ?", quote(column)), value)
}

func (b *Builder) OrWhereNotEq(column string, value any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.OrWhere(fmt.Sprintf("%s != ?", quote(column)), value)
}

func (b *Builder) OrWhereGt(column string, value any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.OrWhere(fmt.Sprintf("%s > ?", quote(column)), value)
}

func (b *Builder) OrWhereLt(column string, value any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.OrWhere(fmt.Sprintf("%s < ?", quote(column)), value)
}

func (b *Builder) OrWhereGe(column string, value any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.OrWhere(fmt.Sprintf("%s >= ?", quote(column)), value)
}

func (b *Builder) OrWhereLe(column string, value any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.OrWhere(fmt.Sprintf("%s <= ?", quote(column)), value)
}

func (b *Builder) OrWhereIn(column string, values []any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}
	if len(values) == 0 {
		return b
	}

	val := make([]string, len(values))
	for i := range values {
		val[i] = "?"
	}

	return b.OrWhere(
		fmt.Sprintf("%s IN (%s)", quote(column), strings.Join(val, ", ")),
		values...)
}

func (b *Builder) OrWhereNotIn(column string, values []any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}
	if len(values) == 0 {
		return b
	}

	val := make([]string, len(values))
	for i := range values {
		val[i] = "?"
	}

	return b.OrWhere(
		fmt.Sprintf("%s NOT IN (%s)", quote(column), strings.Join(val, ", ")),
		values...)
}

func (b *Builder) OrWhereNull(column string) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.OrWhere(fmt.Sprintf("%s IS NULL", quote(column)))
}

func (b *Builder) OrWhereNotNull(column string) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.OrWhere(fmt.Sprintf("%s IS NOT NULL", quote(column)))
}

func (b *Builder) OrWhereBetween(column string, start, end any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.OrWhere(fmt.Sprintf("%s BETWEEN ? AND ?", quote(column)), start, end)
}
