package core

import (
	"fmt"
	"strings"
)

func (b *Builder) buildWhere() string {
	if len(b.WhereList) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(" WHERE ")

	for i, e := range b.WhereList {
		if i > 0 {
			sb.WriteString(" ")
			sb.WriteString(e.Operator)
			sb.WriteString(" ")
		}
		sb.WriteString(e.Condition)
	}

	return sb.String()
}

func (b *Builder) Where(condition string, args ...any) *Builder {
	b.WhereList = append(b.WhereList, Where{
		Condition: condition,
		Operator:  "AND",
	})
	b.WhereArgs = append(b.WhereArgs, args...)
	return b
}

func (b *Builder) WhereEq(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("WhereEq: %w", err))
		return b
	}
	return b.Where(fmt.Sprintf("%s = ?", quote(column)), value)
}

func (b *Builder) WhereNotEq(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("WhereNotEq: %w", err))
		return b
	}
	return b.Where(fmt.Sprintf("%s != ?", quote(column)), value)
}

func (b *Builder) WhereGt(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("WhereGt: %w", err))
		return b
	}
	return b.Where(fmt.Sprintf("%s > ?", quote(column)), value)
}

func (b *Builder) WhereLt(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("WhereLt: %w", err))
		return b
	}
	return b.Where(fmt.Sprintf("%s < ?", quote(column)), value)
}

func (b *Builder) WhereGe(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("WhereGe: %w", err))
		return b
	}
	return b.Where(fmt.Sprintf("%s >= ?", quote(column)), value)
}

func (b *Builder) WhereLe(column string, value any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("WhereLe: %w", err))
		return b
	}
	return b.Where(fmt.Sprintf("%s <= ?", quote(column)), value)
}

func (b *Builder) WhereIn(column string, values []any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("WhereIn: %w", err))
		return b
	}

	if len(values) == 0 {
		b.Error = append(b.Error, fmt.Errorf("WhereIn: values is empty"))
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
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("WhereNotIn: %w", err))
		return b
	}

	if len(values) == 0 {
		b.Error = append(b.Error, fmt.Errorf("WhereNotIn: values is empty"))
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
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("WhereNull: %w", err))
		return b
	}
	return b.Where(fmt.Sprintf("%s IS NULL", quote(column)))
}

func (b *Builder) WhereNotNull(column string) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("WhereNotNull: %w", err))
		return b
	}
	return b.Where(fmt.Sprintf("%s IS NOT NULL", quote(column)))
}

func (b *Builder) WhereBetween(column string, start, end any) *Builder {
	if err := ValidateColumn(column); err != nil {
		b.Error = append(b.Error, fmt.Errorf("WhereBetween: %w", err))
		return b
	}
	return b.Where(fmt.Sprintf("%s BETWEEN ? AND ?", quote(column)), start, end)
}
