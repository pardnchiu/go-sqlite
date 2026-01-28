package core

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
	b.SelectList = columns
	return b
}

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

func (b *Builder) Join(table, on string) *Builder {
	b.JoinList = append(b.JoinList, Join{
		Mode:  "INNER JOIN",
		Table: table,
		On:    on,
	})
	return b
}

func (b *Builder) LeftJoin(table, on string) *Builder {
	b.JoinList = append(b.JoinList, Join{
		Mode:  "LEFT JOIN",
		Table: table,
		On:    on,
	})
	return b
}

func (b *Builder) buildJoin() (string, error) {
	var sb strings.Builder
	for _, e := range b.JoinList {
		if err := ValidateColumn(e.Table); err != nil {
			return "", fmt.Errorf("invalid join table: %w", err)
		}
		if strings.TrimSpace(e.On) == "" {
			return "", fmt.Errorf("join ON clause cannot be empty")
		}
		sb.WriteString(" ")
		sb.WriteString(e.Mode)
		sb.WriteString(" ")
		sb.WriteString(quote(e.Table))
		sb.WriteString(" ON ")
		sb.WriteString(e.On)
	}
	return sb.String(), nil
}

func (b *Builder) OrderBy(column string, direction ...direction) *Builder {
	dir := "ASC"
	if len(direction) > 0 && direction[0] == Desc {
		dir = "DESC"
	}
	b.OrderByList = append(b.OrderByList, fmt.Sprintf("%s %s", quote(column), dir))
	return b
}

func (b *Builder) buildOrderBy() string {
	var sb strings.Builder
	if len(b.OrderByList) == 0 {
		return ""
	}
	sb.WriteString(" ORDER BY ")
	sb.WriteString(strings.Join(b.OrderByList, ", "))
	return sb.String()
}

func (b *Builder) GroupBy(columns ...string) *Builder {
	if len(columns) == 0 {
		return b
	}

	for _, col := range columns {
		if err := ValidateColumn(col); err != nil {
			continue
		}
		b.GroupByList = append(b.GroupByList, col)
	}
	return b
}

func (b *Builder) buildGroupBy() string {
	if len(b.GroupByList) == 0 {
		return ""
	}

	quotedCols := make([]string, len(b.GroupByList))
	for i, col := range b.GroupByList {
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
		b.WithLimit = &num[0]
	} else if len(num) >= 2 {
		b.WithOffset = &num[0]
		b.WithLimit = &num[1]
	}

	return b
}

func (b *Builder) buildLimit() string {
	var sb strings.Builder
	if b.WithLimit == nil {
		return ""
	}
	sb.WriteString(" LIMIT ")
	sb.WriteString(strconv.Itoa(*b.WithLimit))
	return sb.String()
}

func (b *Builder) Offset(num int) *Builder {
	b.WithOffset = &num
	return b
}

func (b *Builder) buildOffset() string {
	var sb strings.Builder
	if b.WithOffset == nil {
		return ""
	}
	sb.WriteString(" OFFSET ")
	sb.WriteString(strconv.Itoa(*b.WithOffset))
	return sb.String()
}

func (b *Builder) Total() *Builder {
	b.WithTotal = true
	return b
}

func (b *Builder) Context(ctx context.Context) *Builder {
	b.WithContext = ctx
	return b
}

func selectBuilder(b *Builder, count bool) (string, error) {
	if b.TableName == nil {
		return "", fmt.Errorf("table name is required")
	}

	if err := ValidateColumn(*b.TableName); err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("SELECT ")

	if count {
		sb.WriteString("COUNT(*)")
	} else if len(b.SelectList) == 0 {
		sb.WriteString("*")
	} else {
		cols := make([]string, len(b.SelectList))
		for i, col := range b.SelectList {
			if col == "*" {
				cols[i] = "*"
			} else {
				if err := ValidateColumn(col); err != nil {
					return "", err
				}
				cols[i] = quote(col)
			}
		}
		sb.WriteString(strings.Join(cols, ", "))
	}

	sb.WriteString(" FROM ")
	sb.WriteString(quote(*b.TableName))

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

	if !count && b.WithTotal {
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

	args := append(b.WhereArgs, b.HavingArgs...)
	if b.WithContext != nil {
		return b.DB.QueryContext(b.WithContext, query, args...)
	}
	return b.DB.Query(query, args...)
}
