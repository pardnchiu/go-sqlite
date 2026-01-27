package goSqlite

import (
	"context"
	"database/sql"
)

func (b *Builder) First() (*sql.Row, error) {
	b.Limit(1)
	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}
	return b.db.QueryRow(query, b.whereArgs...), nil
}

func (b *Builder) FirstContext(ctx context.Context) (*sql.Row, error) {
	b.Limit(1)
	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}
	return b.db.QueryRowContext(ctx, query, b.whereArgs...), nil
}

func (b *Builder) Count() (int64, error) {
	query, err := selectBuilder(b, true)
	if err != nil {
		return 0, err
	}

	var count int64
	err = b.db.QueryRow(query, b.whereArgs...).Scan(&count)
	return count, err
}

func (b *Builder) CountContext(ctx context.Context) (int64, error) {
	query, err := selectBuilder(b, true)
	if err != nil {
		return 0, err
	}

	var count int64
	err = b.db.QueryRowContext(ctx, query, b.whereArgs...).Scan(&count)
	return count, err
}
