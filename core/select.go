package core

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
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

func (b *Builder) Bind(target any) *Builder {
	b.WithBind = target
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

	var targetVal reflect.Value
	var targetElem reflect.Value
	if b.WithBind != nil {
		targetVal = reflect.ValueOf(b.WithBind)
		if targetVal.Kind() != reflect.Pointer {
			return nil, fmt.Errorf("target must be a pointer")
		}

		targetElem = targetVal.Elem()
		switch targetElem.Kind() {
		case reflect.Struct:
			b.Limit(1)
		default:
			break
		}
	}

	rows, err := get(b)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if b.WithBind != nil {
		switch targetElem.Kind() {
		case reflect.Slice:
			return rows, findSlice(rows, targetElem)
		case reflect.Struct:
			return rows, find(rows, targetElem)
		default:
			return rows, fmt.Errorf("target must br struct or slice")
		}
	}
	return rows, err
}

func get(b *Builder) (*sql.Rows, error) {
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

func findSlice(rows *sql.Rows, sliceVal reflect.Value) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	elemType := sliceVal.Type().Elem()

	for rows.Next() {
		item := reflect.New(elemType).Elem()
		scanTarget := scanTarget(item, elemType, cols)

		if err := rows.Scan(scanTarget...); err != nil {
			return err
		}

		sliceVal.Set(reflect.Append(sliceVal, item))
	}

	return rows.Err()
}

func find(rows *sql.Rows, structVal reflect.Value) error {
	if !rows.Next() {
		return sql.ErrNoRows
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	structType := structVal.Type()
	scanTarget := scanTarget(structVal, structType, cols)

	return rows.Scan(scanTarget...)
}

func findRow(row *sql.Row, structVal reflect.Value) error {
	structType := structVal.Type()
	scanDest := make([]any, structType.NumField())

	fieldMap := make(map[string]int)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag != "" {
			fieldMap[dbTag] = i
		} else {
			fieldMap[strings.ToLower(field.Name)] = i
		}
	}

	for i := 0; i < structType.NumField(); i++ {
		scanDest[i] = structVal.Field(i).Addr().Interface()
	}

	return row.Scan(scanDest...)
}

func scanTarget(val reflect.Value, typ reflect.Type, cols []string) []any {
	scanTarget := make([]any, len(cols))

	var dummy any
	for i, col := range cols {
		pattern := getPattern(val, typ, col)
		if pattern == nil {
			scanTarget[i] = &dummy
		} else {
			scanTarget[i] = pattern
		}
	}

	return scanTarget
}

func getPattern(val reflect.Value, typ reflect.Type, colName string) any {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		dbTag := field.Tag.Get("db")

		if dbTag == colName || (dbTag == "" && strings.EqualFold(field.Name, colName)) {
			return val.Field(i).Addr().Interface()
		}
	}

	return nil
}
