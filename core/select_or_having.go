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
	return b.OrHaving(fmt.Sprintf("%s = ?", column), value)
}

func (b *Builder) OrHavingNotEq(column string, value any) *Builder {
	return b.OrHaving(fmt.Sprintf("%s != ?", column), value)
}

func (b *Builder) OrHavingGt(column string, value any) *Builder {
	return b.OrHaving(fmt.Sprintf("%s > ?", column), value)
}

func (b *Builder) OrHavingLt(column string, value any) *Builder {
	return b.OrHaving(fmt.Sprintf("%s < ?", column), value)
}

func (b *Builder) OrHavingGe(column string, value any) *Builder {
	return b.OrHaving(fmt.Sprintf("%s >= ?", column), value)
}

func (b *Builder) OrHavingLe(column string, value any) *Builder {
	return b.OrHaving(fmt.Sprintf("%s <= ?", column), value)
}

func (b *Builder) OrHavingIn(column string, values []any) *Builder {
	if len(values) == 0 {
		return b
	}

	val := make([]string, len(values))
	for i := range values {
		val[i] = "?"
	}

	return b.OrHaving(
		fmt.Sprintf("%s IN (%s)", column, strings.Join(val, ", ")),
		values...)
}

func (b *Builder) OrHavingNotIn(column string, values []any) *Builder {
	if len(values) == 0 {
		return b
	}

	val := make([]string, len(values))
	for i := range values {
		val[i] = "?"
	}

	return b.OrHaving(
		fmt.Sprintf("%s NOT IN (%s)", column, strings.Join(val, ", ")),
		values...)
}

func (b *Builder) OrHavingNull(column string) *Builder {
	return b.OrHaving(fmt.Sprintf("%s IS NULL", column))
}

func (b *Builder) OrHavingNotNull(column string) *Builder {
	return b.OrHaving(fmt.Sprintf("%s IS NOT NULL", column))
}

func (b *Builder) OrHavingBetween(column string, start, end any) *Builder {
	return b.OrHaving(fmt.Sprintf("%s BETWEEN ? AND ?", column), start, end)
}
