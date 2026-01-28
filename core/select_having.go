package core

import (
	"fmt"
	"strings"
)

func (b *Builder) buildHaving() string {
	if len(b.HavingList) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(" HAVING ")

	for i, e := range b.HavingList {
		if i > 0 {
			sb.WriteString(" ")
			sb.WriteString(e.Operator)
			sb.WriteString(" ")
		}
		sb.WriteString(e.Condition)
	}

	return sb.String()
}

func (b *Builder) Having(condition string, args ...any) *Builder {
	b.HavingList = append(b.HavingList, Where{
		Condition: condition,
		Operator:  "AND",
	})
	b.HavingArgs = append(b.HavingArgs, args...)
	return b
}

func (b *Builder) HavingEq(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("HavingEq: %w", err))
		return b
	}
	return b.Having(fmt.Sprintf("%s = ?", quote(column)), value)
}

func (b *Builder) HavingNotEq(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("HavingNotEq: %w", err))
		return b
	}
	return b.Having(fmt.Sprintf("%s != ?", quote(column)), value)
}

func (b *Builder) HavingGt(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("HavingGt: %w", err))
		return b
	}
	return b.Having(fmt.Sprintf("%s > ?", quote(column)), value)
}

func (b *Builder) HavingLt(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("HavingLt: %w", err))
		return b
	}
	return b.Having(fmt.Sprintf("%s < ?", quote(column)), value)
}

func (b *Builder) HavingGe(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("HavingGe: %w", err))
		return b
	}
	return b.Having(fmt.Sprintf("%s >= ?", quote(column)), value)
}

func (b *Builder) HavingLe(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("HavingLe: %w", err))
		return b
	}
	return b.Having(fmt.Sprintf("%s <= ?", quote(column)), value)
}

func (b *Builder) HavingIn(column string, values []any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("HavingIn: %w", err))
		return b
	}

	if len(values) == 0 {
		b.Error = append(b.Error, fmt.Errorf("HavingIn: values is empty"))
		return b
	}

	val := make([]string, len(values))
	for i := range values {
		val[i] = "?"
	}

	return b.Having(
		fmt.Sprintf("%s IN (%s)", quote(column), strings.Join(val, ", ")),
		values...)
}

func (b *Builder) HavingNotIn(column string, values []any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("HavingNotIn: %w", err))
		return b
	}

	if len(values) == 0 {
		b.Error = append(b.Error, fmt.Errorf("HavingNotIn: values is empty"))
		return b
	}

	val := make([]string, len(values))
	for i := range values {
		val[i] = "?"
	}

	return b.Having(
		fmt.Sprintf("%s NOT IN (%s)", quote(column), strings.Join(val, ", ")),
		values...)
}

func (b *Builder) HavingNull(column string) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("HavingNull: %w", err))
		return b
	}
	return b.Having(fmt.Sprintf("%s IS NULL", quote(column)))
}

func (b *Builder) HavingNotNull(column string) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("HavingNotNull: %w", err))
		return b
	}
	return b.Having(fmt.Sprintf("%s IS NOT NULL", quote(column)))
}

func (b *Builder) HavingBetween(column string, start, end any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("HavingBetween: %w", err))
		return b
	}
	return b.Having(fmt.Sprintf("%s BETWEEN ? AND ?", quote(column)), start, end)
}
