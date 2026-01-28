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
	return b.Having(fmt.Sprintf("%s = ?", column), value)
}

func (b *Builder) HavingNotEq(column string, value any) *Builder {
	return b.Having(fmt.Sprintf("%s != ?", column), value)
}

func (b *Builder) HavingGt(column string, value any) *Builder {
	return b.Having(fmt.Sprintf("%s > ?", column), value)
}

func (b *Builder) HavingLt(column string, value any) *Builder {
	return b.Having(fmt.Sprintf("%s < ?", column), value)
}

func (b *Builder) HavingGe(column string, value any) *Builder {
	return b.Having(fmt.Sprintf("%s >= ?", column), value)
}

func (b *Builder) HavingLe(column string, value any) *Builder {
	return b.Having(fmt.Sprintf("%s <= ?", column), value)
}

func (b *Builder) HavingIn(column string, values []any) *Builder {
	if len(values) == 0 {
		return b
	}

	val := make([]string, len(values))
	for i := range values {
		val[i] = "?"
	}

	return b.Having(
		fmt.Sprintf("%s IN (%s)", column, strings.Join(val, ", ")),
		values...)
}

func (b *Builder) HavingNotIn(column string, values []any) *Builder {
	if len(values) == 0 {
		return b
	}

	val := make([]string, len(values))
	for i := range values {
		val[i] = "?"
	}

	return b.Having(
		fmt.Sprintf("%s NOT IN (%s)", column, strings.Join(val, ", ")),
		values...)
}

func (b *Builder) HavingNull(column string) *Builder {
	return b.Having(fmt.Sprintf("%s IS NULL", column))
}

func (b *Builder) HavingNotNull(column string) *Builder {
	return b.Having(fmt.Sprintf("%s IS NOT NULL", column))
}

func (b *Builder) HavingBetween(column string, start, end any) *Builder {
	return b.Having(fmt.Sprintf("%s BETWEEN ? AND ?", column), start, end)
}
