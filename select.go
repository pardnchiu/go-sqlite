package goSqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func (b *Builder) Select(columns ...string) *Builder {
	b.selectList = columns
	return b
}

func (b *Builder) Where(condition string, args ...any) *Builder {
	b.whereList = append(b.whereList, Where{
		condition: condition,
		operator:  "AND",
	})
	b.whereArgs = append(b.whereArgs, args...)
	return b
}

func (b *Builder) OrWhere(condition string, args ...any) *Builder {
	b.whereList = append(b.whereList, Where{
		condition: condition,
		operator:  "OR",
	})
	b.whereArgs = append(b.whereArgs, args...)
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

func (b *Builder) OrderBy(column string, direction ...string) *Builder {
	dir := "ASC"
	if len(direction) > 0 {
		dir = strings.ToUpper(direction[0])
	}
	b.orderBy = append(b.orderBy, fmt.Sprintf("%s %s", quote(column), dir))
	return b
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

func (b *Builder) Offset(num int) *Builder {
	b.offset = &num
	return b
}

func (b *Builder) Total() *Builder {
	b.withTotal = true
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

	for _, e := range b.joinList {
		sb.WriteString(" ")
		sb.WriteString(e.mode)
		sb.WriteString(" ")
		sb.WriteString(quote(e.table))
		sb.WriteString(" ON ")
		sb.WriteString(e.on)
	}

	if !count && b.withTotal {
		query := sb.String()
		sb.Reset()
		sb.WriteString(fmt.Sprintf("SELECT COUNT(*) OVER() AS total, data.* FROM (%s) AS data", query))
	}

	sb.WriteString(b.buildWhere())

	if !count {
		if len(b.orderBy) > 0 {
			sb.WriteString(" ORDER BY ")
			sb.WriteString(strings.Join(b.orderBy, ", "))
		}

		if b.limit != nil {
			sb.WriteString(fmt.Sprintf(" LIMIT %d", *b.limit))
		}

		if b.offset != nil {
			sb.WriteString(fmt.Sprintf(" OFFSET %d", *b.offset))
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
	return b.db.Query(query, b.whereArgs...)
}

func (b *Builder) GetContext(ctx context.Context) (*sql.Rows, error) {
	defer builderClear(b)

	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}
	return b.db.QueryContext(ctx, query, b.whereArgs...)
}
