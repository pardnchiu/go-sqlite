package goSqlite

import (
	"fmt"
	"sort"
	"strings"
)

func (b *Builder) Insert(data map[string]any) error {
	_, err := insert(b, data)
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) InsertReturningID(data map[string]any) (int64, error) {
	id, err := insert(b, data)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func insert(b *Builder, data map[string]any) (int64, error) {
	if b.table == nil {
		return 0, fmt.Errorf("table name is required")
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("no data defined")
	}

	if err := validateColumn(*b.table); err != nil {
		return 0, err
	}

	keys := make([]string, 0, len(data))
	for key := range data {
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
		values = append(values, data[key])
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		quote(*b.table),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := b.db.Exec(query, values...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
