package goSqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	conn *Connector
	once sync.Once
)

func New(c Config) (*Connector, error) {
	once.Do(func() {
		conn = &Connector{db: make(map[string]*sql.DB)}
	})

	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.db == nil {
		conn.db = make(map[string]*sql.DB)
	}

	// get {dbName}.db form path
	if c.Key == "" {
		filename := filepath.Base(c.Path)
		c.Key = strings.TrimSuffix(filename, filepath.Ext(filename))
	}

	if conn.db[c.Key] != nil {
		return conn, nil
	}

	db, err := sql.Open("sqlite3", c.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	db.Exec("PRAGMA journal_mode=WAL")
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(2)

	if c.Lifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(c.Lifetime) * time.Second)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	conn.db[c.Key] = db
	return conn, nil
}

func (d *Connector) DB(key string) (*Builder, error) {
	db, err := db(d, key)
	if err != nil {
		return nil, err
	}
	return NewBuilder(db), nil
}

func db(d *Connector, key string) (*sql.DB, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db[key] == nil {
		return nil, fmt.Errorf("db %s not found", key)
	}
	return d.db[key], nil
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
	d.mu.Lock()
	defer d.mu.Unlock()

	for key, db := range d.db {
		err := db.Close()
		if err == nil {
			continue
		}
		slog.Error("failed to close db",
			slog.String("db", key),
			slog.Any("error", err))
	}
	d.db = nil
}
