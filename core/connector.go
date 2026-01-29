package core

import (
	"context"
	"database/sql"
	"log/slog"
)

func (d *Connector) Query(key, query string, args ...any) (*sql.Rows, error) {
	return d.Read.DB.Query(query, args...)
}

func (d *Connector) QueryContext(ctx context.Context, key, query string, args ...any) (*sql.Rows, error) {
	return d.Read.DB.QueryContext(ctx, query, args...)
}

func (d *Connector) Exec(key, query string, args ...any) (sql.Result, error) {
	return d.Write.DB.Exec(query, args...)
}

func (d *Connector) ExecContext(ctx context.Context, key, query string, args ...any) (sql.Result, error) {
	return d.Write.DB.ExecContext(ctx, query, args...)
}

func (d *Connector) Close() {
	if d.Read != nil && d.Read.DB != nil {
		if err := d.Read.DB.Close(); err != nil {
			slog.Error("failed to close read db",
				slog.Any("error", err))
		}
	}

	if d.Write != nil && d.Write.DB != nil {
		if err := d.Write.DB.Close(); err != nil {
			slog.Error("failed to close write db",
				slog.Any("error", err))
		}
	}
}
