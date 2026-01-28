package goSqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

const (
	Asc direction = iota
	Desc
)

func (b *Builder) Select(columns ...string) *Builder {
	b.selectList = columns
	return b
}

func (b *Builder) buildWhere() string {
	if len(b.whereList) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(" WHERE ")

	for i, e := range b.whereList {
		if i > 0 {
			sb.WriteString(" ")
			sb.WriteString(e.operator)
			sb.WriteString(" ")
		}
		sb.WriteString(e.condition)
	}

	return sb.String()
}

func (b *Builder) Join(table, on string) *Builder {
	b.joinList = append(b.joinList, Join{
		mode:  "INNER JOIN",
		table: table,
		on:    on,
	})
	return b
}

func (b *Builder) LeftJoin(table, on string) *Builder {
	b.joinList = append(b.joinList, Join{
		mode:  "LEFT JOIN",
		table: table,
		on:    on,
	})
	return b
}

func (b *Builder) buildJoin() (string, error) {
	var sb strings.Builder
	for _, e := range b.joinList {
		if err := validateColumn(e.table); err != nil {
			return "", fmt.Errorf("invalid join table: %w", err)
		}
		if strings.TrimSpace(e.on) == "" {
			return "", fmt.Errorf("join ON clause cannot be empty")
		}
		sb.WriteString(" ")
		sb.WriteString(e.mode)
		sb.WriteString(" ")
		sb.WriteString(quote(e.table))
		sb.WriteString(" ON ")
		sb.WriteString(e.on)
	}
	return sb.String(), nil
}

func (b *Builder) OrderBy(column string, direction ...direction) *Builder {
	dir := "ASC"
	if len(direction) > 0 && direction[0] == Desc {
		dir = "DESC"
	}
	b.orderBy = append(b.orderBy, fmt.Sprintf("%s %s", quote(column), dir))
	return b
}

func (b *Builder) buildOrderBy() string {
	var sb strings.Builder
	if len(b.orderBy) == 0 {
		return ""
	}
	sb.WriteString(" ORDER BY ")
	sb.WriteString(strings.Join(b.orderBy, ", "))
	return sb.String()
}

func (b *Builder) GroupBy(columns ...string) *Builder {
	if len(columns) == 0 {
		return b
	}

	for _, col := range columns {
		if err := validateColumn(col); err != nil {
			continue
		}
		b.groupBy = append(b.groupBy, col)
	}
	return b
}

func (b *Builder) buildGroupBy() string {
	if len(b.groupBy) == 0 {
		return ""
	}

	quotedCols := make([]string, len(b.groupBy))
	for i, col := range b.groupBy {
		quotedCols[i] = quote(col)
	}

	var sb strings.Builder
	sb.WriteString(" GROUP BY ")
	sb.WriteString(strings.Join(quotedCols, ", "))
	return sb.String()
}

func (b *Builder) Limit(num ...int) *Builder {
	if len(num) == 0 {
		return b
	}

	if len(num) == 1 {
		b.limit = &num[0]
	} else if len(num) >= 2 {
		b.offset = &num[0]
		b.limit = &num[1]
	}

	return b
}

func (b *Builder) buildLimit() string {
	var sb strings.Builder
	if b.limit == nil {
		return ""
	}
	sb.WriteString(" LIMIT ")
	sb.WriteString(strconv.Itoa(*b.limit))
	return sb.String()
}

func (b *Builder) Offset(num int) *Builder {
	b.offset = &num
	return b
}

func (b *Builder) buildOffset() string {
	var sb strings.Builder
	if b.offset == nil {
		return ""
	}
	sb.WriteString(" OFFSET ")
	sb.WriteString(strconv.Itoa(*b.offset))
	return sb.String()
}

func (b *Builder) Total() *Builder {
	b.withTotal = true
	return b
}

func (b *Builder) Context(ctx context.Context) *Builder {
	b.context = ctx
	return b
}

func selectBuilder(b *Builder, count bool) (string, error) {
	if b.table == nil {
		return "", fmt.Errorf("table name is required")
	}

	if err := validateColumn(*b.table); err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("SELECT ")

	if count {
		sb.WriteString("COUNT(*)")
	} else if len(b.selectList) == 0 {
		sb.WriteString("*")
	} else {
		cols := make([]string, len(b.selectList))
		for i, col := range b.selectList {
			if col == "*" {
				cols[i] = "*"
			} else {
				if err := validateColumn(col); err != nil {
					return "", err
				}
				cols[i] = quote(col)
			}
		}
		sb.WriteString(strings.Join(cols, ", "))
	}

	sb.WriteString(" FROM ")
	sb.WriteString(quote(*b.table))

	query, err := b.buildJoin()
	if err != nil {
		return "", err
	}
	sb.WriteString(query)

	where := b.buildWhere()
	groupBy := b.buildGroupBy()
	having := b.buildHaving()
	orderBy := b.buildOrderBy()
	limit := b.buildLimit()
	offset := b.buildOffset()

	if !count && b.withTotal {
		query := sb.String()

		sb.Reset()
		sb.WriteString("SELECT COUNT(*) OVER() AS total, data.* FROM (")
		sb.WriteString(query)
		sb.WriteString(where)
		sb.WriteString(groupBy)
		sb.WriteString(having)
		sb.WriteString(orderBy)
		sb.WriteString(") AS data")
		sb.WriteString(limit)
		sb.WriteString(offset)
	} else {
		sb.WriteString(where)
		sb.WriteString(groupBy)
		sb.WriteString(having)

		if !count {
			sb.WriteString(orderBy)
			sb.WriteString(limit)
			sb.WriteString(offset)
		}
	}

	return sb.String(), nil
}

func (b *Builder) Get() (*sql.Rows, error) {
	defer builderClear(b)

	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}

	args := append(b.whereArgs, b.havingArgs...)
	if b.context != nil {
		return b.db.QueryContext(b.context, query, args...)
	}
	return b.db.Query(query, args...)
}
