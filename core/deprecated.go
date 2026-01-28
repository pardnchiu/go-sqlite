package core

import (
	"context"
	"database/sql"
)

// ! Deprecated: Use Context(ctx).Insert() in v1.0.0
func (b *Builder) InsertContext(ctx context.Context, data ...map[string]any) (int64, error) {
	defer builderClear(b)

	if len(b.Error) > 0 {
		return 0, b.Error[0]
	}

	query, values, err := insertBuilder(b, data...)
	if err != nil {
		return 0, err
	}

	result, err := b.DB.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ! Deprecated: Use Insert() in v1.0.0
func (b *Builder) InsertReturningID(data ...map[string]any) (int64, error) {
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

// ! Deprecated: Use Context(ctx).Insert() in v1.0.0
func (b *Builder) InsertContextReturningID(ctx context.Context, data ...map[string]any) (int64, error) {
	defer builderClear(b)

	if len(b.Error) > 0 {
		return 0, b.Error[0]
	}

	query, values, err := insertBuilder(b, data...)
	if err != nil {
		return 0, err
	}

	result, err := b.DB.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ! Deprecated: Use Conflict(conflict).Insert() in v1.0.0
func (b *Builder) InsertConflict(conflict conflict, data ...map[string]any) (int64, error) {
	defer builderClear(b)

	if len(b.Error) > 0 {
		return 0, b.Error[0]
	}

	b.ConflictMode = &conflict

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

	if len(b.Error) > 0 {
		return 0, b.Error[0]
	}

	b.ConflictMode = &conflict

	query, values, err := insertBuilder(b, data...)
	if err != nil {
		return 0, err
	}

	result, err := b.DB.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ! Deprecated: Use Conflict(conflict).Insert() in v1.0.0
func (b *Builder) InsertConflictReturningID(conflict conflict, data ...map[string]any) (int64, error) {
	defer builderClear(b)

	if len(b.Error) > 0 {
		return 0, b.Error[0]
	}

	b.ConflictMode = &conflict

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

	if len(b.Error) > 0 {
		return 0, b.Error[0]
	}

	b.ConflictMode = &conflict

	query, values, err := insertBuilder(b, data...)
	if err != nil {
		return 0, err
	}

	result, err := b.DB.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ! Deprecated: Use Context(ctx).Get() in v1.0.0
func (b *Builder) GetContext(ctx context.Context) (*sql.Rows, error) {
	defer builderClear(b)

	if len(b.Error) > 0 {
		return nil, b.Error[0]
	}

	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}
	return b.DB.QueryContext(ctx, query, b.WhereArgs...)
}

// ! Deprecated: Use Total(ctx).Get() in v1.0.0
func (b *Builder) GetWithTotal() (*sql.Rows, error) {
	defer builderClear(b)

	if len(b.Error) > 0 {
		return nil, b.Error[0]
	}

	b.WithTotal = true

	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}

	if b.WithContext != nil {
		return b.DB.QueryContext(b.WithContext, query, b.WhereArgs...)
	}
	return b.DB.Query(query, b.WhereArgs...)
}

// ! Deprecated: Use Total(ctx).Context(ctx).Get() in v1.0.0
func (b *Builder) GetWithTotalContext(ctx context.Context) (*sql.Rows, error) {
	defer builderClear(b)

	if len(b.Error) > 0 {
		return nil, b.Error[0]
	}

	b.WithTotal = true

	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}
	return b.DB.QueryContext(ctx, query, b.WhereArgs...)
}

// ! Deprecated: Use Context(ctx).First() in v1.0.0
func (b *Builder) FirstContext(ctx context.Context) (*sql.Row, error) {
	defer builderClear(b)

	if len(b.Error) > 0 {
		return nil, b.Error[0]
	}

	b.Limit(1)

	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}
	return b.DB.QueryRowContext(ctx, query, b.WhereArgs...), nil
}

// ! Deprecated: Use Context(ctx).Count() in v1.0.0
func (b *Builder) CountContext(ctx context.Context) (int64, error) {
	defer builderClear(b)

	if len(b.Error) > 0 {
		return 0, b.Error[0]
	}

	query, err := selectBuilder(b, true)
	if err != nil {
		return 0, err
	}

	var count int64
	err = b.DB.QueryRowContext(ctx, query, b.WhereArgs...).Scan(&count)
	return count, err
}

// ! Deprecated: Use Context(ctx).Update() in v1.0.0
func (b *Builder) UpdateContext(ctx context.Context, data ...map[string]any) (int64, error) {
	defer builderClear(b)

	if len(b.Error) > 0 {
		return 0, b.Error[0]
	}

	query, values, err := updateBuilder(b, data...)
	if err != nil {
		return 0, err
	}

	result, err := b.DB.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
