package core

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
)

func (d *Connector) DB(key string) (*Builder, error) {
	db, err := db(d, key)
	if err != nil {
		return nil, err
	}
	return NewBuilder(db), nil
}

func db(d *Connector, key string) (*sql.DB, error) {
	d.Mu.Lock()
	defer d.Mu.Unlock()

	if d.Map[key] == nil {
		return nil, fmt.Errorf("db %s not found", key)
	}
	return d.Map[key], nil
}

func (d *Connector) Query(key, query string, args ...any) (*sql.Rows, error) {
	db, err := db(d, key)
	if err != nil {
		return nil, err
	}
	return db.Query(query, args...)
}

func (d *Connector) QueryContext(ctx context.Context, key, query string, args ...any) (*sql.Rows, error) {
	db, err := db(d, key)
	if err != nil {
		return nil, err
	}
	return db.QueryContext(ctx, query, args...)
}

func (d *Connector) Exec(key, query string, args ...any) (sql.Result, error) {
	db, err := db(d, key)
	if err != nil {
		return nil, err
	}
	return db.Exec(query, args...)
}

func (d *Connector) ExecContext(ctx context.Context, key, query string, args ...any) (sql.Result, error) {
	db, err := db(d, key)
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, query, args...)
}

func (d *Connector) Close() {
	d.Mu.Lock()
	defer d.Mu.Unlock()

	for key, db := range d.Map {
		err := db.Close()
		if err == nil {
			continue
		}
		slog.Error("failed to close db",
			slog.String("db", key),
			slog.Any("error", err))
	}
	d.Map = nil
}
