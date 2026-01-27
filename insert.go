package goSqlite

import (
	"fmt"
	"sort"
	"strings"
)

type conflict uint32

const (
	Ignore conflict = iota
	Replace
	Abort
	Fail
	Rollback
)

func (b *Builder) Insert(data ...map[string]any) error {
	_, err := insert(b, nil, data...)
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) InsertReturningID(data ...map[string]any) (int64, error) {
	id, err := insert(b, nil, data...)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (b *Builder) InsertConflict(conflict conflict, data ...map[string]any) error {
	_, err := insert(b, &conflict, data...)
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) InsertConflictReturningID(conflict conflict, data ...map[string]any) (int64, error) {
	id, err := insert(b, &conflict, data...)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func insert(b *Builder, conflict *conflict, data ...map[string]any) (int64, error) {
	if b.table == nil {
		return 0, fmt.Errorf("table name is required")
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("no data defined")
	}

	if err := validateColumn(*b.table); err != nil {
		return 0, err
	}

	insertData := data[0]
	var conflictData map[string]any
	if len(data) > 1 {
		conflictData = data[1]
	}

	keys := make([]string, 0, len(insertData))
	for key := range insertData {
		if err := validateColumn(key); err != nil {
			return 0, err
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
	if conflict != nil {
		sb.WriteString(" OR ")
		switch *conflict {
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
	sb.WriteString(quote(*b.table))
	sb.WriteString(" (")
	sb.WriteString(strings.Join(columns, ", "))
	sb.WriteString(") VALUES (")
	sb.WriteString(strings.Join(placeholders, ", "))
	sb.WriteString(")")

	if conflictData != nil && len(conflictData) > 0 {
		updateKeys := make([]string, 0, len(conflictData))
		for key := range conflictData {
			if err := validateColumn(key); err != nil {
				return 0, err
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

	result, err := b.db.Exec(sb.String(), values...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
