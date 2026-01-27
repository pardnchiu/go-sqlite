package goSqlite

import (
	"fmt"
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

	columns := make([]string, 0, len(data))
	values := make([]any, 0, len(data))
	placeholders := make([]string, 0, len(data))

	for column, value := range data {
		columns = append(columns, fmt.Sprintf("`%s`", column))
		values = append(values, value)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
		*b.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := b.db.Exec(query, values...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
