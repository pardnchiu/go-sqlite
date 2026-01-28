package goSqlite

import (
	"context"
	"database/sql"
)

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

// ! Deprecated: Use Context(ctx).Get() in v1.0.0
func (b *Builder) GetContext(ctx context.Context) (*sql.Rows, error) {
	defer builderClear(b)

	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}
	return b.db.QueryContext(ctx, query, b.whereArgs...)
}

// ! Deprecated: Use Total(ctx).Get() in v1.0.0
func (b *Builder) GetWithTotal() (*sql.Rows, error) {
	defer builderClear(b)

	b.withTotal = true

	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}

	if b.context != nil {
		return b.db.QueryContext(b.context, query, b.whereArgs...)
	}
	return b.db.Query(query, b.whereArgs...)
}

// ! Deprecated: Use Total(ctx).Context(ctx).Get() in v1.0.0
func (b *Builder) GetWithTotalContext(ctx context.Context) (*sql.Rows, error) {
	defer builderClear(b)

	b.withTotal = true

	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}
	return b.db.QueryContext(ctx, query, b.whereArgs...)
}

// ! Deprecated: Use Context(ctx).First() in v1.0.0
func (b *Builder) FirstContext(ctx context.Context) (*sql.Row, error) {
	b.Limit(1)
	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}
	return b.db.QueryRowContext(ctx, query, b.whereArgs...), nil
}

// ! Deprecated: Use Context(ctx).Count() in v1.0.0
func (b *Builder) CountContext(ctx context.Context) (int64, error) {
	query, err := selectBuilder(b, true)
	if err != nil {
		return 0, err
	}

	var count int64
	err = b.db.QueryRowContext(ctx, query, b.whereArgs...).Scan(&count)
	return count, err
}

// ! Deprecated: Use Context(ctx).Update() in v1.0.0
func (b *Builder) UpdateContext(ctx context.Context, data ...map[string]any) (int64, error) {
	defer builderClear(b)

	query, values, err := updateBuilder(b, data...)
	if err != nil {
		return 0, err
	}

	result, err := b.db.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
