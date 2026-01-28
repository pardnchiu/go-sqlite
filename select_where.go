package goSqlite

import (
	"fmt"
	"strings"
)

func (b *Builder) Where(condition string, args ...any) *Builder {
	b.whereList = append(b.whereList, Where{
		condition: condition,
		operator:  "AND",
	})
	b.whereArgs = append(b.whereArgs, args...)
	return b
}

func (b *Builder) WhereEq(column string, value any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.Where(fmt.Sprintf("%s = ?", quote(column)), value)
}

func (b *Builder) WhereNotEq(column string, value any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.Where(fmt.Sprintf("%s != ?", quote(column)), value)
}

func (b *Builder) WhereGt(column string, value any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.Where(fmt.Sprintf("%s > ?", quote(column)), value)
}

func (b *Builder) WhereLt(column string, value any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.Where(fmt.Sprintf("%s < ?", quote(column)), value)
}

func (b *Builder) WhereGe(column string, value any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.Where(fmt.Sprintf("%s >= ?", quote(column)), value)
}

func (b *Builder) WhereLe(column string, value any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.Where(fmt.Sprintf("%s <= ?", quote(column)), value)
}

func (b *Builder) WhereIn(column string, values []any) *Builder {
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

	return b.Where(
		fmt.Sprintf("%s IN (%s)", quote(column), strings.Join(val, ", ")),
		values...)
}

func (b *Builder) WhereNotIn(column string, values []any) *Builder {
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

	return b.Where(
		fmt.Sprintf("%s NOT IN (%s)", quote(column), strings.Join(val, ", ")),
		values...)
}

func (b *Builder) WhereNull(column string) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.Where(fmt.Sprintf("%s IS NULL", quote(column)))
}

func (b *Builder) WhereNotNull(column string) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.Where(fmt.Sprintf("%s IS NOT NULL", quote(column)))
}

func (b *Builder) WhereBetween(column string, start, end any) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	return b.Where(fmt.Sprintf("%s BETWEEN ? AND ?", quote(column)), start, end)
}
