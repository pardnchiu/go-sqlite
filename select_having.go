package goSqlite

import (
	"fmt"
	"strings"
)

func (b *Builder) buildHaving() string {
	if len(b.havingList) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(" HAVING ")

	for i, e := range b.havingList {
		if i > 0 {
			sb.WriteString(" ")
			sb.WriteString(e.operator)
			sb.WriteString(" ")
		}
		sb.WriteString(e.condition)
	}

	return sb.String()
}

func (b *Builder) Having(condition string, args ...any) *Builder {
	b.havingList = append(b.havingList, Where{
		condition: condition,
		operator:  "AND",
	})
	b.havingArgs = append(b.havingArgs, args...)
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
