package goSqlite

import (
	"database/sql"
)

func (b *Builder) First() (*sql.Row, error) {
	b.Limit(1)
	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}

	args := append(b.whereArgs, b.havingArgs...)
	var row *sql.Row
	if b.context != nil {
		row = b.db.QueryRowContext(b.context, query, args...)
	} else {
		row = b.db.QueryRow(query, args...)
	}
	return row, nil
}

func (b *Builder) Count() (int64, error) {
	query, err := selectBuilder(b, true)
	if err != nil {
		return 0, err
	}

	args := append(b.whereArgs, b.havingArgs...)
	var count int64
	if b.context != nil {
		err = b.db.QueryRowContext(b.context, query, args...).Scan(&count)
	} else {
		err = b.db.QueryRow(query, args...).Scan(&count)
	}
	return count, err
}
