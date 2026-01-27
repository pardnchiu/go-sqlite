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

type Database struct {
	db map[string]*sql.DB
	mu sync.Mutex
}

var (
	database *Database
)

func New(c Config) (*Database, *sql.DB, error) {
	if database == nil {
		database = &Database{db: make(map[string]*sql.DB)}
	}

	database.mu.Lock()
	defer database.mu.Unlock()

	// get {dbName}.db form path
	if c.Key == "" {
		filename := filepath.Base(c.Path)
		c.Key = strings.TrimSuffix(filename, filepath.Ext(filename))
	}

	if database.db == nil {
		database.db = make(map[string]*sql.DB)
	}

	if database.db[c.Key] != nil {
		return database, database.db[c.Key], nil
	}

	db, err := sql.Open("sqlite3", c.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if c.Lifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(c.Lifetime) * time.Second)
	}

	if err := db.Ping(); err != nil {
		return nil, nil, fmt.Errorf("failed to ping db: %w", err)
	}

	database.db[c.Key] = db
	return database, db, nil
}

func (d *Database) DB(key string) (*Builder, error) {
	db, err := db(d, key)
	if err != nil {
		return nil, err
	}
	return NewBuilder(db), nil
}

func db(d *Database, key string) (*sql.DB, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db[key] == nil {
		return nil, fmt.Errorf("db %s not found", key)
	}
	return d.db[key], nil
}

func (d *Database) Query(key, query string, args ...any) (*sql.Rows, error) {
	db, err := db(d, key)
	if err != nil {
		return nil, err
	}
	return db.Query(query, args...)
}

func (d *Database) QueryContext(ctx context.Context, key, query string, args ...any) (*sql.Rows, error) {
	db, err := db(d, key)
	if err != nil {
		return nil, err
	}
	return db.QueryContext(ctx, query, args...)
}

func (d *Database) Exec(key, query string, args ...any) (sql.Result, error) {
	db, err := db(d, key)
	if err != nil {
		return nil, err
	}
	return db.Exec(query, args...)
}

func (d *Database) ExecContext(ctx context.Context, key, query string, args ...any) (sql.Result, error) {
	db, err := db(d, key)
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, query, args...)
}

func (d *Database) Close() {
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
