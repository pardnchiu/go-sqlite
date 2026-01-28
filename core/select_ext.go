package core

import (
	"database/sql"
)

func (b *Builder) First() (*sql.Row, error) {
	b.Limit(1)
	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}

	args := append(b.WhereArgs, b.HavingArgs...)
	var row *sql.Row
	if b.WithContext != nil {
		row = b.DB.QueryRowContext(b.WithContext, query, args...)
	} else {
		row = b.DB.QueryRow(query, args...)
	}
	return row, nil
}

func (b *Builder) Count() (int64, error) {
	query, err := selectBuilder(b, true)
	if err != nil {
		return 0, err
	}

	args := append(b.WhereArgs, b.HavingArgs...)
	var count int64
	if b.WithContext != nil {
		err = b.DB.QueryRowContext(b.WithContext, query, args...).Scan(&count)
	} else {
		err = b.DB.QueryRow(query, args...).Scan(&count)
	}
	return count, err
}
