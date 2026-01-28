package core

import (
	"fmt"
	"strings"
)

func (b *Builder) OrHaving(condition string, args ...any) *Builder {
	b.HavingList = append(b.HavingList, Where{
		Condition: condition,
		Operator:  "OR",
	})
	b.HavingArgs = append(b.HavingArgs, args...)
	return b
}

func (b *Builder) OrHavingEq(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrHavingEq: %w", err))
		return b
	}
	return b.OrHaving(fmt.Sprintf("%s = ?", quote(column)), value)
}

func (b *Builder) OrHavingNotEq(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrHavingNotEq: %w", err))
		return b
	}
	return b.OrHaving(fmt.Sprintf("%s != ?", quote(column)), value)
}

func (b *Builder) OrHavingGt(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrHavingGt: %w", err))
		return b
	}
	return b.OrHaving(fmt.Sprintf("%s > ?", quote(column)), value)
}

func (b *Builder) OrHavingLt(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrHavingLt: %w", err))
		return b
	}
	return b.OrHaving(fmt.Sprintf("%s < ?", quote(column)), value)
}

func (b *Builder) OrHavingGe(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrHavingGe: %w", err))
		return b
	}
	return b.OrHaving(fmt.Sprintf("%s >= ?", quote(column)), value)
}

func (b *Builder) OrHavingLe(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrHavingLe: %w", err))
		return b
	}
	return b.OrHaving(fmt.Sprintf("%s <= ?", quote(column)), value)
}

func (b *Builder) OrHavingIn(column string, values []any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrHavingIn: %w", err))
		return b
	}

	if len(values) == 0 {
		b.Error = append(b.Error, fmt.Errorf("OrHavingIn: values is empty"))
		return b
	}

	val := make([]string, len(values))
	for i := range values {
		val[i] = "?"
	}

	return b.OrHaving(
		fmt.Sprintf("%s IN (%s)", quote(column), strings.Join(val, ", ")),
		values...)
}

func (b *Builder) OrHavingNotIn(column string, values []any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrHavingNotIn: %w", err))
		return b
	}

	if len(values) == 0 {
		b.Error = append(b.Error, fmt.Errorf("OrHavingNotIn: values is empty"))
		return b
	}

	val := make([]string, len(values))
	for i := range values {
		val[i] = "?"
	}

	return b.OrHaving(
		fmt.Sprintf("%s NOT IN (%s)", quote(column), strings.Join(val, ", ")),
		values...)
}

func (b *Builder) OrHavingNull(column string) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrHavingNull: %w", err))
		return b
	}
	return b.OrHaving(fmt.Sprintf("%s IS NULL", quote(column)))
}

func (b *Builder) OrHavingNotNull(column string) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrHavingNotNull: %w", err))
		return b
	}
	return b.OrHaving(fmt.Sprintf("%s IS NOT NULL", quote(column)))
}

func (b *Builder) OrHavingBetween(column string, start, end any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("OrHavingBetween: %w", err))
		return b
	}
	return b.OrHaving(fmt.Sprintf("%s BETWEEN ? AND ?", quote(column)), start, end)
}
