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

type Config struct {
	Key      string `json:"key,omitempty"`
	Path     string `json:"path"`
	Lifetime int    `json:"lifetime,omitempty"` // sec
}

type Connector struct {
	db map[string]*sql.DB
	mu sync.Mutex
}

var (
	Conn *Connector
)

func New(c Config) (*Connector, error) {
	if Conn == nil {
		Conn = &Connector{db: make(map[string]*sql.DB)}
	}

	Conn.mu.Lock()
	defer Conn.mu.Unlock()

	// get {dbName}.db form path
	if c.Key == "" {
		filename := filepath.Base(c.Path)
		c.Key = strings.TrimSuffix(filename, filepath.Ext(filename))
	}

	if Conn.db == nil {
		Conn.db = make(map[string]*sql.DB)
	}

	if Conn.db[c.Key] != nil {
		return Conn, nil
	}

	db, err := sql.Open("sqlite3", c.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if c.Lifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(c.Lifetime) * time.Second)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	Conn.db[c.Key] = db
	return Conn, nil
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
