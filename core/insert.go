package core

import (
	"fmt"
	"sort"
	"strings"
)

const (
	Ignore conflict = iota
	Replace
	Abort
	Fail
	Rollback
)

func (b *Builder) Conflict(conflict conflict) *Builder {
	b.ConflictMode = &conflict
	return b
}

func (b *Builder) Insert(data ...map[string]any) (int64, error) {
	defer builderClear(b)

	if len(b.Error) > 0 {
		return 0, b.Error[0]
	}

	query, values, err := insertBuilder(b, data...)
	if err != nil {
		return 0, err
	}

	result, err := b.ExecAutoAsignContext(query, values...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func insertBuilder(b *Builder, data ...map[string]any) (string, []any, error) {
	if b.TableName == nil {
		return "", nil, fmt.Errorf("table name is required")
	}

	if len(data) == 0 {
		return "", nil, fmt.Errorf("no data defined")
	}

	if err := ValidateColumn(*b.TableName); err != nil {
		return "", nil, err
	}

	insertData := data[0]
	var conflictData map[string]any
	if len(data) > 1 {
		conflictData = data[1]
	}

	keys := make([]string, 0, len(insertData))
	for key := range insertData {
		if err := ValidateColumn(key); err != nil {
			return "", nil, err
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	columns := make([]string, 0, len(keys))
	values := make([]any, 0, len(keys))
	placeholders := make([]string, 0, len(keys))

	for _, key := range keys {
		columns = append(columns, quote(key))
		values = append(values, insertData[key])
		placeholders = append(placeholders, "?")
	}

	var sb strings.Builder
	sb.WriteString("INSERT")
	if b.ConflictMode != nil {
		sb.WriteString(" OR ")
		switch *b.ConflictMode {
		case Ignore:
			sb.WriteString("IGNORE")
		case Replace:
			sb.WriteString("REPLACE")
		case Abort:
			sb.WriteString("ABORT")
		case Fail:
			sb.WriteString("FAIL")
		case Rollback:
			sb.WriteString("ROLLBACK")
		}
	}

	sb.WriteString(" INTO ")
	sb.WriteString(quote(*b.TableName))
	sb.WriteString(" (")
	sb.WriteString(strings.Join(columns, ", "))
	sb.WriteString(") VALUES (")
	sb.WriteString(strings.Join(placeholders, ", "))
	sb.WriteString(")")

	if conflictData != nil && len(conflictData) > 0 {
		updateKeys := make([]string, 0, len(conflictData))
		for key := range conflictData {
			if err := ValidateColumn(key); err != nil {
				return "", nil, err
			}
			updateKeys = append(updateKeys, key)
		}
		sort.Strings(updateKeys)

		setParts := make([]string, 0, len(updateKeys))
		for _, key := range updateKeys {
			setParts = append(setParts, fmt.Sprintf("%s = ?", quote(key)))
			values = append(values, conflictData[key])
		}

		sb.WriteString(" ON CONFLICT DO UPDATE SET ")
		sb.WriteString(strings.Join(setParts, ", "))
	}

	return sb.String(), values, nil
}

func (b *Builder) InsertBatch(data []map[string]any) (int64, error) {
	defer builderClear(b)

	if len(b.Error) > 0 {
		return 0, b.Error[0]
	}

	if len(data) == 0 {
		return 0, fmt.Errorf("no data to insert")
	}

	query, values, err := insertBatchBuilder(b, data)
	if err != nil {
		return 0, err
	}

	result, err := b.ExecAutoAsignContext(query, values...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func insertBatchBuilder(b *Builder, data []map[string]any) (string, []any, error) {
	if b.TableName == nil {
		return "", nil, fmt.Errorf("table name is required")
	}

	if len(data) == 0 {
		return "", nil, fmt.Errorf("no data defined")
	}

	if err := ValidateColumn(*b.TableName); err != nil {
		return "", nil, err
	}

	insertData := data[0]
	keys := make([]string, 0, len(insertData))
	for key := range insertData {
		if err := ValidateColumn(key); err != nil {
			return "", nil, err
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(quote(*b.TableName))
	sb.WriteString(" (")

	quotedKeys := make([]string, len(keys))
	for i, key := range keys {
		quotedKeys[i] = quote(key)
	}
	sb.WriteString(strings.Join(quotedKeys, ", "))
	sb.WriteString(") VALUES ")

	values := make([]any, 0, len(data)*len(keys))
	for i, row := range data {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString("(")
		placeholders := make([]string, len(keys))
		for j, key := range keys {
			placeholders[j] = "?"
			values = append(values, row[key])
		}
		sb.WriteString(strings.Join(placeholders, ", "))
		sb.WriteString(")")
	}

	return sb.String(), values, nil
}
