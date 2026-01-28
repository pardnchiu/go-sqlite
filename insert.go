package goSqlite

import (
	"context"
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

func (b *Builder) Conflict(conflict conflict) *Builder {
	b.conflict = &conflict
	return b
}

func (b *Builder) Insert(data ...map[string]any) (int64, error) {
	defer builderClear(b)

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

// ! Deprecated: Use Context(ctx).Insert() in v1.0.0
func (b *Builder) InsertContext(ctx context.Context, data ...map[string]any) (int64, error) {
	defer builderClear(b)

	query, values, err := insertBuilder(b, data...)
	if err != nil {
		return 0, err
	}

	result, err := b.db.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ! Deprecated: Use Insert() in v1.0.0
func (b *Builder) InsertReturningID(data ...map[string]any) (int64, error) {
	defer builderClear(b)

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

// ! Deprecated: Use Context(ctx).Insert() in v1.0.0
func (b *Builder) InsertContextReturningID(ctx context.Context, data ...map[string]any) (int64, error) {
	defer builderClear(b)

	query, values, err := insertBuilder(b, data...)
	if err != nil {
		return 0, err
	}

	result, err := b.db.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ! Deprecated: Use Conflict(conflict).Insert() in v1.0.0
func (b *Builder) InsertConflict(conflict conflict, data ...map[string]any) (int64, error) {
	defer builderClear(b)

	b.conflict = &conflict

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

// ! Deprecated: Use Conflict(conflict).Context(ctx).Insert() in v1.0.0
func (b *Builder) InsertContexConflict(ctx context.Context, conflict conflict, data ...map[string]any) (int64, error) {
	defer builderClear(b)

	b.conflict = &conflict

	query, values, err := insertBuilder(b, data...)
	if err != nil {
		return 0, err
	}

	result, err := b.db.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ! Deprecated: Use Conflict(conflict).Insert() in v1.0.0
func (b *Builder) InsertConflictReturningID(conflict conflict, data ...map[string]any) (int64, error) {
	defer builderClear(b)

	b.conflict = &conflict

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

// ! Deprecated: Use Conflict(conflict).Context(ctx).Insert() in v1.0.0
func (b *Builder) InsertContextConflictReturningID(ctx context.Context, conflict conflict, data ...map[string]any) (int64, error) {
	defer builderClear(b)

	b.conflict = &conflict

	query, values, err := insertBuilder(b, data...)
	if err != nil {
		return 0, err
	}

	result, err := b.db.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func insertBuilder(b *Builder, data ...map[string]any) (string, []any, error) {
	if b.table == nil {
		return "", []any{}, fmt.Errorf("table name is required")
	}
	if len(data) == 0 {
		return "", []any{}, fmt.Errorf("no data defined")
	}

	if err := validateColumn(*b.table); err != nil {
		return "", []any{}, err
	}

	insertData := data[0]
	var conflictData map[string]any
	if len(data) > 1 {
		conflictData = data[1]
	}

	keys := make([]string, 0, len(insertData))
	for key := range insertData {
		if err := validateColumn(key); err != nil {
			return "", []any{}, err
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
	if b.conflict != nil {
		sb.WriteString(" OR ")
		switch *b.conflict {
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
				return "", []any{}, err
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
