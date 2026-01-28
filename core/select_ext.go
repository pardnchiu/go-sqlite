package core

import (
	"database/sql"
)

func (b *Builder) First() (*sql.Row, error) {
	defer builderClear(b)

	b.Limit(1)

	if len(b.Error) > 0 {
		return nil, b.Error[0]
	}

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
	defer builderClear(b)

	if len(b.Error) > 0 {
		return 0, b.Error[0]
	}

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
