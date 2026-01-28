// update.go 新增

package goSqlite

import (
	"fmt"
	"sort"
	"strings"
)

func (b *Builder) Increase(column string, num ...int) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	n := 1
	if len(num) > 0 {
		n = num[0]
	}

	b.updateList = append(b.updateList, fmt.Sprintf("%s = %s + %d", quote(column), quote(column), n))
	return b
}

func (b *Builder) Decrease(column string, num ...int) *Builder {
	if err := validateColumn(column); err != nil {
		return b
	}

	n := 1
	if len(num) > 0 {
		n = num[0]
	}

	b.updateList = append(b.updateList, fmt.Sprintf("%s = %s - %d", quote(column), quote(column), n))
	return b
}

func (b *Builder) Toggle(column string) *Builder {

	if err := validateColumn(column); err != nil {
		return b
	}
	b.updateList = append(b.updateList, fmt.Sprintf("%s = NOT %s", quote(column), quote(column)))
	return b
}

func (b *Builder) Update(data ...map[string]any) (int64, error) {
	defer builderClear(b)

	query, values, err := updateBuilder(b, data...)
	if err != nil {
		return 0, err
	}

	result, err := b.ExecAutoAsignContext(query, values...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func updateBuilder(b *Builder, data ...map[string]any) (string, []any, error) {
	if b.table == nil {
		return "", []any{}, fmt.Errorf("table name is required")
	}

	if err := validateColumn(*b.table); err != nil {
		return "", []any{}, err
	}

	var mainData map[string]any
	if len(data) > 0 {
		mainData = data[0]
	}

	if mainData == nil && len(b.updateList) == 0 {
		return "", nil, fmt.Errorf("no data defined")
	}

	var sb strings.Builder
	sb.WriteString("UPDATE ")
	sb.WriteString(quote(*b.table))
	sb.WriteString(" SET ")

	parts := make([]string, 0)
	values := make([]any, 0)

	if len(b.updateList) > 0 {
		parts = append(parts, b.updateList...)
	}

	if len(data) > 0 {
		keys := make([]string, 0, len(mainData))
		for key := range mainData {
			if err := validateColumn(key); err != nil {
				return "", []any{}, err
			}
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s = ?", quote(k)))
			values = append(values, mainData[k])
		}
	}

	sb.WriteString(strings.Join(parts, ", "))
	sb.WriteString(b.buildWhere())

	values = append(values, b.whereArgs...)

	return sb.String(), values, nil
}
