package core

import (
	"fmt"
	"strings"
)

func (b *Builder) OrWhere(condition string, args ...any) *Builder {
	b.WhereList = append(b.WhereList, Where{
		Condition: condition,
		Operator:  "OR",
	})
	b.WhereArgs = append(b.WhereArgs, args...)
	return b
}

func (b *Builder) OrWhereEq(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrWhereEq: %w", err))
		return b
	}
	return b.OrWhere(fmt.Sprintf("%s = ?", quote(column)), value)
}

func (b *Builder) OrWhereNotEq(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrWhereNotEq: %w", err))
		return b
	}
	return b.OrWhere(fmt.Sprintf("%s != ?", quote(column)), value)
}

func (b *Builder) OrWhereGt(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrWhereGt: %w", err))
		return b
	}
	return b.OrWhere(fmt.Sprintf("%s > ?", quote(column)), value)
}

func (b *Builder) OrWhereLt(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrWhereLt: %w", err))
		return b
	}
	return b.OrWhere(fmt.Sprintf("%s < ?", quote(column)), value)
}

func (b *Builder) OrWhereGe(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrWhereGe: %w", err))
		return b
	}
	return b.OrWhere(fmt.Sprintf("%s >= ?", quote(column)), value)
}

func (b *Builder) OrWhereLe(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrWhereLe: %w", err))
		return b
	}
	return b.OrWhere(fmt.Sprintf("%s <= ?", quote(column)), value)
}

func (b *Builder) OrWhereIn(column string, values []any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrWhereIn: %w", err))
		return b
	}

	if len(values) == 0 {
		b.Error = append(b.Error, fmt.Errorf("OrWhereIn: values is empty"))
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
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrWhereNotIn: %w", err))
		return b
	}

	if len(values) == 0 {
		b.Error = append(b.Error, fmt.Errorf("OrWhereNotIn: values is empty"))
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
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrWhereNull: %w", err))
		return b
	}
	return b.OrWhere(fmt.Sprintf("%s IS NULL", quote(column)))
}

func (b *Builder) OrWhereNotNull(column string) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrWhereNotNull: %w", err))
		return b
	}
	return b.OrWhere(fmt.Sprintf("%s IS NOT NULL", quote(column)))
}

func (b *Builder) OrWhereBetween(column string, start, end any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrWhereBetween: %w", err))
		return b
	}
	return b.OrWhere(fmt.Sprintf("%s BETWEEN ? AND ?", quote(column)), start, end)
}
